package app

import (
	"context"
	"crypt-server/internal/store"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := TemplateData{
		Title: "Crypt",
		User:  s.currentUser(r),
	}
	computers, err := s.store.ListComputers()
	if err != nil {
		s.renderError(w, err)
		return
	}
	outstanding, err := s.store.ListOutstandingRequests()
	if err != nil {
		s.renderError(w, err)
		return
	}
	data.Computers = computers
	data.OutstandingCount = len(outstanding)

	if err := s.renderTemplate(w, r, "index", data); err != nil {
		s.renderError(w, err)
	}
}

func (s *Server) handleTableAjax(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	draw := 0
	if raw := r.URL.Query().Get("args"); raw != "" {
		var payload map[string]any
		if err := json.Unmarshal([]byte(raw), &payload); err == nil {
			if value, ok := payload["draw"].(float64); ok {
				draw = int(value)
			}
		}
	}

	computers, err := s.store.ListComputers()
	if err != nil {
		s.renderError(w, err)
		return
	}
	data["draw"] = draw
	data["recordsTotal"] = len(computers)
	data["recordsFiltered"] = len(computers)

	rows := make([][]string, 0, len(computers))
	for _, computer := range computers {
		serial := html.EscapeString(computer.Serial)
		computerName := html.EscapeString(computer.ComputerName)
		username := html.EscapeString(computer.Username)
		lastCheckin := ""
		if !computer.LastCheckin.IsZero() {
			lastCheckin = computer.LastCheckin.Format("2006-01-02 15:04")
		}

		link := fmt.Sprintf("/info/%d/", computer.ID)
		rows = append(rows, []string{
			fmt.Sprintf("<a href=\"%s\">%s</a>", link, serial),
			fmt.Sprintf("<a href=\"%s\">%s</a>", link, computerName),
			username,
			lastCheckin,
			fmt.Sprintf("<a class=\"btn btn-info btn-xs\" href=\"%s\">Info</a>", link),
		})
	}

	data["data"] = rows

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.renderError(w, err)
	}
}

func (s *Server) handleNewComputer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := TemplateData{Title: "New Computer", User: s.currentUser(r)}
		if err := s.renderTemplate(w, r, "new_computer", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		serial := strings.TrimSpace(r.FormValue("serial"))
		username := strings.TrimSpace(r.FormValue("username"))
		computerName := strings.TrimSpace(r.FormValue("computername"))
		if serial == "" || computerName == "" {
			data := TemplateData{
				Title:        "New Computer",
				User:         s.currentUser(r),
				ErrorMessage: "Serial number and computer name are required.",
			}
			if err := s.renderTemplate(w, r, "new_computer", data); err != nil {
				s.renderError(w, err)
			}
			return
		}
		computer, err := s.store.AddComputer(serial, username, computerName)
		if err != nil {
			s.renderError(w, err)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/info/%d/", computer.ID), http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleNewSecret(w http.ResponseWriter, r *http.Request) {
	computerID, err := idFromPath("/new/secret/", r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	computer, err := s.store.GetComputerByID(computerID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		data := TemplateData{Title: "New Secret", User: s.currentUser(r), Computer: computer}
		if err := s.renderTemplate(w, r, "new_secret", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		secretType := strings.TrimSpace(r.FormValue("secret_type"))
		secret := strings.TrimSpace(r.FormValue("secret"))
		rotationRequired := r.FormValue("rotation_required") == "on"

		if secretType == "" || secret == "" {
			data := TemplateData{
				Title:        "New Secret",
				User:         s.currentUser(r),
				Computer:     computer,
				ErrorMessage: "Secret type and value are required.",
			}
			if err := s.renderTemplate(w, r, "new_secret", data); err != nil {
				s.renderError(w, err)
			}
			return
		}

		_, newSecretEscrowed, err := s.store.AddSecret(computer.ID, secretType, secret, rotationRequired)
		if err != nil {
			s.renderError(w, err)
			return
		}
		user := s.currentUser(r)
		if newSecretEscrowed {
			s.logger.Printf("secret escrowed manually: serial=%s type=%s by_user=%s", computer.Serial, secretType, user.Username)
		} else {
			s.logger.Printf("secret updated manually: serial=%s type=%s by_user=%s", computer.Serial, secretType, user.Username)
		}
		http.Redirect(w, r, fmt.Sprintf("/info/%d/", computer.ID), http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleComputerInfo(w http.ResponseWriter, r *http.Request) {
	identifier := strings.TrimPrefix(r.URL.Path, "/info/")
	identifier = strings.TrimSuffix(identifier, "/")
	if identifier == "" {
		http.NotFound(w, r)
		return
	}

	computer, err := s.lookupComputer(identifier)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	secrets, err := s.store.ListSecretsByComputer(computer.ID)
	if err != nil {
		s.renderError(w, err)
		return
	}

	data := TemplateData{
		Title:    "Computer Info",
		User:     s.currentUser(r),
		Computer: computer,
		Secrets:  secrets,
	}
	if err := s.renderTemplate(w, r, "computer_info", data); err != nil {
		s.renderError(w, err)
	}
}

func (s *Server) handleSecretInfo(w http.ResponseWriter, r *http.Request) {
	secretID, err := idFromPath("/info/secret/", r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	secret, err := s.store.GetSecretByID(secretID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	computer, err := s.store.GetComputerByID(secret.ComputerID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	requests, err := s.store.ListRequestsBySecret(secret.ID)
	if err != nil {
		s.renderError(w, err)
		return
	}
	canRequest := true
	for _, request := range requests {
		if request.RequestingUser == s.currentUser(r).Username && request.Approved == nil {
			canRequest = false
		}
	}
	approved := false
	approvedRequestID := 0
	for _, request := range requests {
		if request.RequestingUser == s.currentUser(r).Username && request.Approved != nil && *request.Approved {
			approved = true
			approvedRequestID = request.ID
		}
	}

	data := TemplateData{
		Title:             "Secret Info",
		User:              s.currentUser(r),
		Computer:          computer,
		Secret:            secret,
		CanRequest:        canRequest,
		RequestApproved:   approved,
		ApprovedRequestID: approvedRequestID,
		RequestsForSecret: requests,
	}
	if err := s.renderTemplate(w, r, "secret_info", data); err != nil {
		s.renderError(w, err)
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	secretID, err := idFromPath("/request/", r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	secret, err := s.store.GetSecretByID(secretID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	computer, err := s.store.GetComputerByID(secret.ComputerID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		data := TemplateData{Title: "Request Secret", User: s.currentUser(r), Secret: secret, Computer: computer}
		if err := s.renderTemplate(w, r, "request", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		reason := strings.TrimSpace(r.FormValue("reason_for_request"))
		user := s.currentUser(r)
		var approved *bool
		var approver string
		if user.CanApprove && s.settings.ApproveOwn {
			approvedValue := true
			approved = &approvedValue
			approver = user.Username
		}
		_, err := s.store.AddRequest(secret.ID, user.Username, reason, approver, approved)
		if err != nil {
			s.renderError(w, err)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/info/secret/%d/", secret.ID), http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	requestID, err := idFromPath("/approve/", r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	req, err := s.store.GetRequestByID(requestID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	secret, err := s.store.GetSecretByID(req.SecretID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	computer, err := s.store.GetComputerByID(secret.ComputerID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if !s.canApproveRequest(r, req) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		data := TemplateData{Title: "Approve Request", User: s.currentUser(r), Request: req, Secret: secret, Computer: computer}
		if err := s.renderTemplate(w, r, "approve", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if !s.canApproveRequest(r, req) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		approvedValue := r.FormValue("approved") == "1"
		reason := strings.TrimSpace(r.FormValue("reason_for_approval"))
		if _, err := s.store.ApproveRequest(req.ID, approvedValue, reason, s.currentUser(r).Username); err != nil {
			s.renderError(w, err)
			return
		}
		http.Redirect(w, r, "/manage-requests/", http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	requestID, err := idFromPath("/retrieve/", r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	req, err := s.store.GetRequestByID(requestID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if req.Approved == nil || !*req.Approved {
		http.Error(w, "request not approved", http.StatusForbidden)
		return
	}
	user := s.currentUser(r)
	if user.Username != req.RequestingUser && !user.CanApprove {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	secret, err := s.store.GetSecretByID(req.SecretID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if s.settings.RotateViewedSecrets {
		updated, err := s.store.SetSecretRotationRequired(secret.ID, true)
		if err != nil {
			s.renderError(w, err)
			return
		}
		secret = updated
	}

	computer, err := s.store.GetComputerByID(secret.ComputerID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	s.logger.Printf("secret retrieved: serial=%s type=%s by_user=%s requested_by=%s", computer.Serial, secret.SecretType, user.Username, req.RequestingUser)

	secretChars := make([]SecretChar, 0, len(secret.Secret))
	for _, char := range secret.Secret {
		entry := SecretChar{Char: string(char), Class: "other"}
		if unicode.IsLetter(char) {
			entry.Class = "letter"
		} else if unicode.IsDigit(char) {
			entry.Class = "number"
		}
		secretChars = append(secretChars, entry)
	}

	data := TemplateData{
		Title:       "Retrieve Secret",
		User:        s.currentUser(r),
		Request:     req,
		Secret:      secret,
		Computer:    computer,
		SecretChars: secretChars,
	}
	if err := s.renderTemplate(w, r, "retrieve", data); err != nil {
		s.renderError(w, err)
	}
}

func (s *Server) handleManageRequests(w http.ResponseWriter, r *http.Request) {
	if !s.currentUser(r).CanApprove {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	requests, err := s.store.ListOutstandingRequests()
	if err != nil {
		s.renderError(w, err)
		return
	}
	views := make([]RequestView, 0, len(requests))
	for _, req := range requests {
		secret, err := s.store.GetSecretByID(req.SecretID)
		if err != nil {
			continue
		}
		computer, err := s.store.GetComputerByID(secret.ComputerID)
		if err != nil {
			continue
		}
		views = append(views, RequestView{
			ID:               req.ID,
			Serial:           computer.Serial,
			ComputerName:     computer.ComputerName,
			RequestingUser:   req.RequestingUser,
			ReasonForRequest: req.ReasonForRequest,
			DateRequested:    req.DateRequested.Format(store.DateTimeFormat),
		})
	}
	data := TemplateData{Title: "Manage Requests", User: s.currentUser(r), ManageRequests: views}
	if err := s.renderTemplate(w, r, "manage_requests", data); err != nil {
		s.renderError(w, err)
	}
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/users/" {
		s.handleUserList(w, r)
		return
	}
	if r.URL.Path == "/admin/users/new/" {
		s.handleNewUser(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/edit/") {
		s.handleUserEdit(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/password/") {
		s.handleUserPassword(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/delete/") {
		s.handleUserDelete(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) handleUserList(w http.ResponseWriter, r *http.Request) {
	if !s.currentUser(r).IsStaff {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	users, err := s.store.ListUsers()
	if err != nil {
		s.renderError(w, err)
		return
	}
	data := TemplateData{Title: "Users", User: s.currentUser(r), Users: users}
	if err := s.renderTemplate(w, r, "user_list", data); err != nil {
		s.renderError(w, err)
	}
}

func (s *Server) handleNewUser(w http.ResponseWriter, r *http.Request) {
	if !s.currentUser(r).IsStaff {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	switch r.Method {
	case http.MethodGet:
		data := TemplateData{
			Title: "New User",
			User:  s.currentUser(r),
			NewUser: UserForm{
				LocalLoginEnabled: true,
				AuthSource:        "local",
			},
		}
		if err := s.renderTemplate(w, r, "user_new", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		isStaff := r.FormValue("is_staff") == "on"
		canApprove := r.FormValue("can_approve") == "on"
		localLoginEnabled := r.FormValue("local_login_enabled") == "on"
		mustReset := r.FormValue("must_reset_password") == "on"
		authSource := strings.TrimSpace(r.FormValue("auth_source"))
		if authSource == "" {
			authSource = "local"
		}
		if username == "" || (localLoginEnabled && password == "") {
			data := TemplateData{
				Title:        "New User",
				User:         s.currentUser(r),
				ErrorMessage: "Username and password are required when local login is enabled.",
				NewUser: UserForm{
					Username:          username,
					IsStaff:           isStaff,
					CanApprove:        canApprove,
					LocalLoginEnabled: localLoginEnabled,
					MustResetPassword: mustReset,
					AuthSource:        authSource,
				},
			}
			if err := s.renderTemplate(w, r, "user_new", data); err != nil {
				s.renderError(w, err)
			}
			return
		}
		if _, err := s.store.GetUserByUsername(username); err == nil {
			data := TemplateData{
				Title:        "New User",
				User:         s.currentUser(r),
				ErrorMessage: "Username already exists.",
				NewUser: UserForm{
					Username:          username,
					IsStaff:           isStaff,
					CanApprove:        canApprove,
					LocalLoginEnabled: localLoginEnabled,
					MustResetPassword: mustReset,
					AuthSource:        authSource,
				},
			}
			if err := s.renderTemplate(w, r, "user_new", data); err != nil {
				s.renderError(w, err)
			}
			return
		} else if err != store.ErrNotFound {
			s.renderError(w, err)
			return
		}
		var passwordHash string
		if localLoginEnabled {
			var err error
			passwordHash, err = hashPassword(password)
			if err != nil {
				s.renderError(w, err)
				return
			}
		}
		if _, err := s.store.AddUser(username, passwordHash, isStaff, canApprove, localLoginEnabled, mustReset, authSource); err != nil {
			s.renderError(w, err)
			return
		}
		s.logger.Printf("user created: username=%s is_staff=%t can_approve=%t by_user=%s", username, isStaff, canApprove, s.currentUser(r).Username)
		if _, err := s.store.AddAuditEvent(s.currentUser(r).Username, username, "user_created", "", clientIP(r)); err != nil {
			s.renderError(w, err)
			return
		}
		http.Redirect(w, r, "/admin/users/", http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleUserEdit(w http.ResponseWriter, r *http.Request) {
	if !s.currentUser(r).IsStaff {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	userID, err := idFromPath("/admin/users/", strings.TrimSuffix(r.URL.Path, "/edit/")+"/")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		data := TemplateData{
			Title:     "Edit User",
			User:      s.currentUser(r),
			AdminUser: user,
			NewUser: UserForm{
				Username:          user.Username,
				IsStaff:           user.IsStaff,
				CanApprove:        user.CanApprove,
				LocalLoginEnabled: user.LocalLoginEnabled,
				MustResetPassword: user.MustResetPassword,
				AuthSource:        user.AuthSource,
			},
		}
		if err := s.renderTemplate(w, r, "user_edit", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		isStaff := r.FormValue("is_staff") == "on"
		canApprove := r.FormValue("can_approve") == "on"
		localLoginEnabled := r.FormValue("local_login_enabled") == "on"
		mustReset := r.FormValue("must_reset_password") == "on"
		authSource := strings.TrimSpace(r.FormValue("auth_source"))
		if authSource == "" {
			authSource = "local"
		}
		if username == "" {
			data := TemplateData{
				Title:        "Edit User",
				User:         s.currentUser(r),
				AdminUser:    user,
				ErrorMessage: "Username is required.",
				NewUser: UserForm{
					Username:          username,
					IsStaff:           isStaff,
					CanApprove:        canApprove,
					LocalLoginEnabled: localLoginEnabled,
					MustResetPassword: mustReset,
					AuthSource:        authSource,
				},
			}
			if err := s.renderTemplate(w, r, "user_edit", data); err != nil {
				s.renderError(w, err)
			}
			return
		}
		if existing, err := s.store.GetUserByUsername(username); err == nil && existing.ID != user.ID {
			data := TemplateData{
				Title:        "Edit User",
				User:         s.currentUser(r),
				AdminUser:    user,
				ErrorMessage: "Username already exists.",
				NewUser: UserForm{
					Username:          username,
					IsStaff:           isStaff,
					CanApprove:        canApprove,
					LocalLoginEnabled: localLoginEnabled,
					MustResetPassword: mustReset,
					AuthSource:        authSource,
				},
			}
			if err := s.renderTemplate(w, r, "user_edit", data); err != nil {
				s.renderError(w, err)
			}
			return
		} else if err != nil && err != store.ErrNotFound {
			s.renderError(w, err)
			return
		}
		updated, err := s.store.UpdateUser(user.ID, username, isStaff, canApprove, localLoginEnabled, mustReset, authSource)
		if err != nil {
			s.renderError(w, err)
			return
		}
		s.logger.Printf("user edited: username=%s is_staff=%t can_approve=%t by_user=%s", username, isStaff, canApprove, s.currentUser(r).Username)
		if reason := buildUserUpdateReason(user, updated); reason != "" {
			if _, err := s.store.AddAuditEvent(s.currentUser(r).Username, updated.Username, "user_updated", reason, clientIP(r)); err != nil {
				s.renderError(w, err)
				return
			}
		}
		if user.MustResetPassword != mustReset {
			action := "force_reset_disabled"
			if mustReset {
				action = "force_reset_enabled"
			}
			if _, err := s.store.AddAuditEvent(s.currentUser(r).Username, updated.Username, action, "", clientIP(r)); err != nil {
				s.renderError(w, err)
				return
			}
		}
		http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit/", updated.ID), http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleUserPassword(w http.ResponseWriter, r *http.Request) {
	if !s.currentUser(r).IsStaff {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	userID, err := idFromPath("/admin/users/", strings.TrimSuffix(r.URL.Path, "/password/")+"/")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		data := TemplateData{Title: "Reset Password", User: s.currentUser(r), AdminUser: user}
		if err := s.renderTemplate(w, r, "user_password", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		password := r.FormValue("password")
		if password == "" {
			data := TemplateData{
				Title:        "Reset Password",
				User:         s.currentUser(r),
				AdminUser:    user,
				ErrorMessage: "Password is required.",
			}
			if err := s.renderTemplate(w, r, "user_password", data); err != nil {
				s.renderError(w, err)
			}
			return
		}
		passwordHash, err := hashPassword(password)
		if err != nil {
			s.renderError(w, err)
			return
		}
		if _, err := s.store.UpdateUserPassword(user.ID, passwordHash, false); err != nil {
			s.renderError(w, err)
			return
		}
		if _, err := s.store.AddAuditEvent(s.currentUser(r).Username, user.Username, "password_reset", "", clientIP(r)); err != nil {
			s.renderError(w, err)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit/", user.ID), http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAuditLog(w http.ResponseWriter, r *http.Request) {
	if !s.currentUser(r).IsStaff {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	page := parsePage(r.URL.Query().Get("page"))
	const pageSize = 50
	offset := (page - 1) * pageSize

	var (
		events []*store.AuditEvent
		total  int
		err    error
	)
	if query == "" {
		total, err = s.store.CountAuditEvents()
		if err != nil {
			s.renderError(w, err)
			return
		}
		events, err = s.store.ListAuditEventsPaged(pageSize, offset)
	} else {
		total, err = s.store.CountSearchAuditEvents(query)
		if err != nil {
			s.renderError(w, err)
			return
		}
		events, err = s.store.SearchAuditEventsPaged(query, pageSize, offset)
	}
	if err != nil {
		s.renderError(w, err)
		return
	}

	totalPages := 1
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
		if page > totalPages {
			page = totalPages
		}
	}

	data := TemplateData{
		Title:           "Audit Log",
		User:            s.currentUser(r),
		AuditEvents:     events,
		AuditSearch:     query,
		AuditPage:       page,
		AuditPageSize:   pageSize,
		AuditTotal:      total,
		AuditTotalPages: totalPages,
	}
	if err := s.renderTemplate(w, r, "audit_log", data); err != nil {
		s.renderError(w, err)
	}
}

func (s *Server) handleUserDelete(w http.ResponseWriter, r *http.Request) {
	if !s.currentUser(r).IsStaff {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	userID, err := idFromPath("/admin/users/", strings.TrimSuffix(r.URL.Path, "/delete/")+"/")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		data := TemplateData{Title: "Delete User", User: s.currentUser(r), AdminUser: user}
		if err := s.renderTemplate(w, r, "user_delete", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if s.currentUser(r).ID == user.ID {
			data := TemplateData{
				Title:        "Delete User",
				User:         s.currentUser(r),
				AdminUser:    user,
				ErrorMessage: "You cannot delete your own account.",
			}
			if err := s.renderTemplate(w, r, "user_delete", data); err != nil {
				s.renderError(w, err)
			}
			return
		}
		if err := s.store.DeleteUser(user.ID); err != nil {
			s.renderError(w, err)
			return
		}
		s.logger.Printf("user deleted: username=%s by_user=%s", user.Username, s.currentUser(r).Username)
		if _, err := s.store.AddAuditEvent(s.currentUser(r).Username, user.Username, "user_deleted", "", clientIP(r)); err != nil {
			s.renderError(w, err)
			return
		}
		http.Redirect(w, r, "/admin/users/", http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePasswordChange(w http.ResponseWriter, r *http.Request) {
	user := s.currentUser(r)
	if !user.LocalLoginEnabled {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	switch r.Method {
	case http.MethodGet:
		data := TemplateData{
			Title:                         "Change Password",
			User:                          user,
			PasswordChangeRequiresCurrent: true,
		}
		if err := s.renderTemplate(w, r, "password_change", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		current := r.FormValue("current_password")
		next := r.FormValue("new_password")
		dbUser, err := s.store.GetUserByUsername(user.Username)
		if err != nil || dbUser.PasswordHash == "" || !verifyPassword(current, dbUser.PasswordHash) {
			data := TemplateData{
				Title:                         "Change Password",
				User:                          user,
				ErrorMessage:                  "Current password is incorrect.",
				PasswordChangeRequiresCurrent: true,
			}
			if err := s.renderTemplate(w, r, "password_change", data); err != nil {
				s.renderError(w, err)
			}
			return
		}
		hash, err := hashPassword(next)
		if err != nil {
			s.renderError(w, err)
			return
		}
		if _, err := s.store.UpdateUserPassword(user.ID, hash, false); err != nil {
			s.renderError(w, err)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePasswordReset(w http.ResponseWriter, r *http.Request) {
	user := s.currentUser(r)
	if !user.LocalLoginEnabled {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if !user.MustResetPassword {
		http.Redirect(w, r, "/password/change/", http.StatusSeeOther)
		return
	}
	switch r.Method {
	case http.MethodGet:
		data := TemplateData{
			Title:                         "Reset Password",
			User:                          user,
			PasswordChangeRequiresCurrent: false,
		}
		if err := s.renderTemplate(w, r, "password_change", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		next := r.FormValue("new_password")
		hash, err := hashPassword(next)
		if err != nil {
			s.renderError(w, err)
			return
		}
		if _, err := s.store.UpdateUserPassword(user.ID, hash, false); err != nil {
			s.renderError(w, err)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSAMLLogin(w http.ResponseWriter, r *http.Request) {
	if s.samlSP == nil {
		http.Error(w, "SAML login not configured", http.StatusNotImplemented)
		return
	}
	s.samlSP.HandleStartAuthFlow(w, r)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := TemplateData{Title: "Login", User: s.currentUser(r)}
		if err := s.renderTemplate(w, r, "login", data); err != nil {
			s.renderError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderError(w, err)
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		next := r.FormValue("next")

		user, err := s.store.GetUserByUsername(username)
		if err != nil || !user.LocalLoginEnabled || user.PasswordHash == "" {
			s.renderLoginError(w, r, "Invalid username or password.")
			return
		}
		if !verifyPassword(password, user.PasswordHash) {
			s.renderLoginError(w, r, "Invalid username or password.")
			return
		}
		token, err := s.sessionManager.Create(user.Username)
		if err != nil {
			s.renderError(w, err)
			return
		}
		s.sessionManager.SetCookie(w, token, s.settings.CookieSecure)
		if user.MustResetPassword && user.LocalLoginEnabled {
			next = "/password/reset/"
		}
		if next == "" || !strings.HasPrefix(next, "/") {
			next = "/"
		}
		http.Redirect(w, r, next, http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.sessionManager.ClearCookie(w, s.settings.CookieSecure)
	http.Redirect(w, r, "/login/", http.StatusSeeOther)
}

func (s *Server) handleCheckin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	serial := r.FormValue("serial")
	if serial == "" {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	recoveryPass := r.FormValue("recovery_password")
	if recoveryPass == "" {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	userName := r.FormValue("username")
	if userName == "" {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	macName := r.FormValue("macname")
	if macName == "" {
		macName = serial
	}
	secretType := r.FormValue("secret_type")
	if secretType == "" {
		secretType = "recovery_key"
	}
	secretType = strings.TrimSpace(secretType)
	if secretType == "" {
		secretType = "recovery_key"
	}

	now := time.Now()
	computer, err := s.store.UpsertComputer(serial, userName, macName, now)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	secret, newSecretEscrowed, err := s.store.AddSecret(computer.ID, secretType, recoveryPass, false)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if newSecretEscrowed {
		s.logger.Printf("secret escrowed: serial=%s type=%s username=%s macname=%s", serial, secretType, userName, macName)
	} else {
		s.logger.Printf("secret updated: serial=%s type=%s username=%s macname=%s", serial, secretType, userName, macName)
	}

	payload := map[string]any{
		"serial":              computer.Serial,
		"username":            computer.Username,
		"rotation_required":   secret.RotationRequired,
		"new_secret_escrowed": newSecretEscrowed,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/verify/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	serial := parts[0]
	secretType := parts[1]
	if serial == "" || secretType == "" {
		http.NotFound(w, r)
		return
	}

	computer, err := s.store.GetComputerBySerial(serial)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	secret, err := s.store.GetLatestSecretByComputerAndType(computer.ID, secretType)
	payload := map[string]any{}
	if err == nil {
		payload["escrowed"] = true
		payload["date_escrowed"] = secret.DateEscrowed.Format(time.RFC3339)
	} else if err == store.ErrNotFound {
		payload["escrowed"] = false
	} else {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) currentUser(r *http.Request) User {
	if user := userFromContext(r.Context()); user != nil {
		return *user
	}
	return User{}
}

func (s *Server) renderError(w http.ResponseWriter, err error) {
	s.logger.Printf("handler error: %v", err)
	http.Error(w, "Something went wrong", http.StatusInternalServerError)
}

func (s *Server) renderTemplate(w http.ResponseWriter, r *http.Request, name string, data TemplateData) error {
	data.CSRFToken = s.csrfToken(w, r)
	data.SAMLAvailable = s.samlSP != nil
	data.SAMLLoginURL = s.samlLoginURL()
	data.Version = Version
	return s.renderer.Render(w, name, data)
}

func buildUserUpdateReason(before, after *store.User) string {
	changes := make([]string, 0, 5)
	if before.Username != after.Username {
		changes = append(changes, "username")
	}
	if before.IsStaff != after.IsStaff {
		changes = append(changes, "is_staff")
	}
	if before.CanApprove != after.CanApprove {
		changes = append(changes, "can_approve")
	}
	if before.LocalLoginEnabled != after.LocalLoginEnabled {
		changes = append(changes, "local_login_enabled")
	}
	if before.AuthSource != after.AuthSource {
		changes = append(changes, "auth_source")
	}
	return strings.Join(changes, ",")
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func parsePage(raw string) int {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 1
	}
	page, err := strconv.Atoi(value)
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func (s *Server) csrfToken(w http.ResponseWriter, r *http.Request) string {
	if s.csrfManager == nil {
		return ""
	}
	token, err := s.csrfManager.EnsureToken(w, r, s.settings.CookieSecure)
	if err != nil {
		return ""
	}
	return token
}

func (s *Server) samlLoginURL() string {
	if s.samlConfig == nil {
		return "/saml/login/"
	}
	if strings.HasPrefix(s.samlConfig.MetadataURLPath, "/saml2/") {
		return "/saml2/login/"
	}
	return "/saml/login/"
}

func (s *Server) renderLoginError(w http.ResponseWriter, r *http.Request, message string) {
	data := TemplateData{
		Title:        "Login",
		User:         s.currentUser(r),
		ErrorMessage: message,
	}
	if err := s.renderTemplate(w, r, "login", data); err != nil {
		s.renderError(w, err)
	}
}

func idFromPath(prefix, path string) (int, error) {
	if !strings.HasPrefix(path, prefix) {
		return 0, errors.New("invalid path")
	}
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return 0, errors.New("missing id")
	}
	return strconv.Atoi(trimmed)
}

func (s *Server) lookupComputer(identifier string) (*store.Computer, error) {
	if id, err := strconv.Atoi(identifier); err == nil {
		return s.store.GetComputerByID(id)
	}
	return s.store.GetComputerBySerial(identifier)
}

func (s *Server) loadUserFromRequest(r *http.Request) *User {
	if s.sessionManager == nil {
		return s.loadUserFromSAML(r)
	}
	cookie, err := r.Cookie(s.sessionManager.CookieName())
	if err != nil {
		return s.loadUserFromSAML(r)
	}
	username, ok := s.sessionManager.Validate(cookie.Value)
	if !ok {
		return s.loadUserFromSAML(r)
	}
	dbUser, err := s.store.GetUserByUsername(username)
	if err != nil {
		return s.loadUserFromSAML(r)
	}
	user := mapStoreUser(dbUser)
	user.IsAuthenticated = true
	if s.settings.AllApprove {
		user.CanApprove = true
	}
	return &user
}

func (s *Server) loadUserFromSAML(r *http.Request) *User {
	if s.samlSP == nil || s.samlConfig == nil {
		return nil
	}
	session, err := s.samlSP.Session.GetSession(r)
	if err != nil {
		return nil
	}
	username := usernameFromSAML(session, s.samlConfig)
	if username == "" {
		return nil
	}
	attributes := attributesFromSession(session)
	groups := groupMembership(attributes, s.samlConfig.GroupsAttribute)
	isStaff, canApprove := resolveSAMLPermissions(groups, s.samlConfig)

	dbUser, err := s.store.GetUserByUsername(username)
	if err != nil {
		if err != store.ErrNotFound || !s.samlConfig.CreateUnknownUser {
			return nil
		}
		newUser, err := s.store.AddUser(
			username,
			"",
			isStaff,
			canApprove,
			s.samlConfig.DefaultLocalLogin,
			s.samlConfig.DefaultMustReset,
			s.samlConfig.DefaultAuthSource,
		)
		if err != nil {
			return nil
		}
		s.logger.Printf("user created via SAML: username=%s is_staff=%t can_approve=%t", username, isStaff, canApprove)
		_, _ = s.store.AddAuditEvent("saml", username, "user_created", "created via SAML login", clientIP(r))
		user := mapStoreUser(newUser)
		user.IsAuthenticated = true
		return &user
	}

	updatedUser := dbUser
	shouldUpdate := false
	if s.shouldUpdateStaff() {
		if dbUser.IsStaff != isStaff {
			updatedUser.IsStaff = isStaff
			shouldUpdate = true
		}
	}
	if s.shouldUpdateApprover() {
		if dbUser.CanApprove != canApprove {
			updatedUser.CanApprove = canApprove
			shouldUpdate = true
		}
	}
	if s.samlConfig.DefaultAuthSource != "" && dbUser.AuthSource != s.samlConfig.DefaultAuthSource {
		updatedUser.AuthSource = s.samlConfig.DefaultAuthSource
		shouldUpdate = true
	}
	if shouldUpdate {
		updatedUser, err = s.store.UpdateUser(
			dbUser.ID,
			dbUser.Username,
			updatedUser.IsStaff,
			updatedUser.CanApprove,
			dbUser.LocalLoginEnabled,
			dbUser.MustResetPassword,
			updatedUser.AuthSource,
		)
		if err != nil {
			return nil
		}
		s.logger.Printf("user updated via SAML: username=%s is_staff=%t can_approve=%t", username, updatedUser.IsStaff, updatedUser.CanApprove)
		_, _ = s.store.AddAuditEvent("saml", username, "user_updated", "permissions updated via SAML login", clientIP(r))
	}
	user := mapStoreUser(updatedUser)
	user.IsAuthenticated = true
	return &user
}

func (s *Server) shouldUpdateStaff() bool {
	if s.samlConfig == nil {
		return false
	}
	return len(s.samlConfig.StaffGroups) > 0 || len(s.samlConfig.SuperuserGroups) > 0
}

func (s *Server) shouldUpdateApprover() bool {
	if s.samlConfig == nil {
		return false
	}
	return len(s.samlConfig.CanApproveGroups) > 0 || len(s.samlConfig.SuperuserGroups) > 0
}

func mapStoreUser(user *store.User) User {
	return User{
		ID:                user.ID,
		Username:          user.Username,
		IsStaff:           user.IsStaff,
		CanApprove:        user.CanApprove,
		LocalLoginEnabled: user.LocalLoginEnabled,
		MustResetPassword: user.MustResetPassword,
		AuthSource:        user.AuthSource,
	}
}

func userFromContext(ctx context.Context) *User {
	if value := ctx.Value(userContextKey); value != nil {
		if user, ok := value.(*User); ok {
			return user
		}
	}
	return nil
}

func (s *Server) canApproveRequest(r *http.Request, req *store.Request) bool {
	user := s.currentUser(r)
	if !user.CanApprove {
		return false
	}
	if !s.settings.ApproveOwn && user.Username == req.RequestingUser {
		return false
	}
	return true
}
