package app

import (
	"context"
	"crypt-server/internal/store"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/crewjam/saml/samlsp"
)

// Version is the application version displayed in the UI.
// Set at build time via: go build -ldflags "-X crypt-server/internal/app.Version=1.0.0"
var Version = "0.0.0-dev"

type Server struct {
	store          store.Store
	renderer       *Renderer
	logger         *log.Logger
	sessionManager *SessionManager
	csrfManager    *CSRFManager
	samlSP         *samlsp.Middleware
	samlConfig     *SAMLConfig
	settings       Settings
}

func NewServer(store store.Store, renderer *Renderer, logger *log.Logger, sessionManager *SessionManager, csrfManager *CSRFManager, samlSP *samlsp.Middleware, samlConfig *SAMLConfig, settings Settings) *Server {
	server := &Server{
		store:          store,
		renderer:       renderer,
		logger:         logger,
		sessionManager: sessionManager,
		csrfManager:    csrfManager,
		samlSP:         samlSP,
		samlConfig:     samlConfig,
		settings:       settings,
	}
	server.startRequestCleanupJob()
	return server
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/login/", s.handleLogin)
	mux.HandleFunc("/logout/", s.handleLogout)
	mux.HandleFunc("/saml/login/", s.handleSAMLLogin)
	mux.HandleFunc("/saml2/login/", s.handleSAMLLogin)
	if s.samlSP != nil {
		mux.Handle("/saml/", s.samlSP)
		mux.Handle("/saml2/", s.samlSP)
	}
	mux.HandleFunc("/checkin/", s.handleCheckin)
	mux.HandleFunc("/verify/", s.handleVerify)
	mux.HandleFunc("/", s.requireAuth(s.handleIndex))
	mux.HandleFunc("/ajax/", s.requireAuth(s.handleTableAjax))
	mux.HandleFunc("/new/computer/", s.requireAuth(s.handleNewComputer))
	mux.HandleFunc("/new/secret/", s.requireAuth(s.handleNewSecret))
	mux.HandleFunc("/info/secret/", s.requireAuth(s.handleSecretInfo))
	mux.HandleFunc("/info/", s.requireAuth(s.handleComputerInfo))
	mux.HandleFunc("/request/", s.requireAuth(s.handleRequest))
	mux.HandleFunc("/retrieve/", s.requireAuth(s.handleRetrieve))
	mux.HandleFunc("/approve/", s.requireAuth(s.handleApprove))
	mux.HandleFunc("/manage-requests/", s.requireAuth(s.handleManageRequests))
	mux.HandleFunc("/admin/users/", s.requireAuth(s.handleAdminUsers))
	mux.HandleFunc("/admin/audit/", s.requireAuth(s.handleAuditLog))
	mux.HandleFunc("/password/change/", s.requireAuth(s.handlePasswordChange))
	mux.HandleFunc("/password/reset/", s.requireAuth(s.handlePasswordReset))

	return withTrailingSlashRedirect(s.withCSRF(s.withUser(mux)))
}

func withTrailingSlashRedirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't redirect static files
		if !strings.HasPrefix(r.URL.Path, "/static/") && r.URL.Path != "/" && !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := s.currentUser(r)
		if !user.IsAuthenticated {
			http.Redirect(w, r, "/login/?next="+urlQueryEscape(r.URL.Path), http.StatusSeeOther)
			return
		}
		if user.MustResetPassword && user.LocalLoginEnabled && r.URL.Path != "/password/reset/" && r.URL.Path != "/logout/" {
			http.Redirect(w, r, "/password/reset/", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func urlQueryEscape(value string) string {
	return url.QueryEscape(value)
}

type contextKey string

const userContextKey contextKey = "user"

func (s *Server) withCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.csrfManager == nil {
			next.ServeHTTP(w, r)
			return
		}
		if !s.isCSRFExempt(r) && isUnsafeMethod(r.Method) {
			if !s.csrfManager.ValidateRequest(r) {
				http.Error(w, "CSRF token missing or invalid", http.StatusForbidden)
				return
			}
		}
		if _, err := s.csrfManager.EnsureToken(w, r, s.settings.CookieSecure); err != nil {
			s.logger.Printf("csrf error: %v", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func (s *Server) isCSRFExempt(r *http.Request) bool {
	switch {
	case strings.HasPrefix(r.URL.Path, "/checkin/"):
		return true
	case strings.HasPrefix(r.URL.Path, "/verify/"):
		return true
	case strings.HasPrefix(r.URL.Path, "/saml/"):
		return true
	case strings.HasPrefix(r.URL.Path, "/saml2/"):
		return true
	default:
		return false
	}
}

func (s *Server) withUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.loadUserFromRequest(r)
		if user != nil {
			ctx := contextWithUser(r.Context(), user)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func contextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
