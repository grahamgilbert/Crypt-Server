package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"crypt-server/internal/crypto"
	"github.com/fernet/fernet-go"
)

func parseFixture(data []byte) ([]fixtureEntry, error) {
	var entries []fixtureEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func convertFixture(entries []fixtureEntry, legacyKey *fernet.Key, newCodec *crypto.AesGcmCodec, passwordMap map[string]passwordMapEntry) (*migrationOutput, error) {
	users := make(map[int]userOut)
	usernames := make(map[int]string)
	emails := make(map[int]string)
	groups := make(map[int]string)
	permissionIDs := make(map[int]struct{})
	userPermissions := make(map[int][]int)
	groupPermissions := make(map[int][]int)
	userGroups := make(map[int][]int)
	computers := make([]computerOut, 0)
	secrets := make([]secretOut, 0)
	requests := make([]requestOut, 0)

	for _, entry := range entries {
		pk := entry.pkInt()
		switch entry.Model {
		case "auth.user":
			username := getString(entry.Fields, "username")
			user := userOut{
				ID:       pk,
				Username: username,
				Email:    getString(entry.Fields, "email"),
				IsStaff:  getBool(entry.Fields, "is_staff"),
				IsSuper:  getBool(entry.Fields, "is_superuser"),
				Groups:   []string{},
			}
			users[pk] = user
			usernames[pk] = username
			emails[pk] = strings.ToLower(getString(entry.Fields, "email"))
		case "auth.group":
			groups[pk] = getString(entry.Fields, "name")
		case "auth.permission":
			if getString(entry.Fields, "codename") == "can_approve" {
				permissionIDs[pk] = struct{}{}
			}
		case "auth.user_user_permissions":
			userID := getInt(entry.Fields, "user")
			permID := getInt(entry.Fields, "permission")
			userPermissions[userID] = append(userPermissions[userID], permID)
		case "auth.group_permissions":
			groupID := getInt(entry.Fields, "group")
			permID := getInt(entry.Fields, "permission")
			groupPermissions[groupID] = append(groupPermissions[groupID], permID)
		case "auth.user_groups":
			userID := getInt(entry.Fields, "user")
			groupID := getInt(entry.Fields, "group")
			userGroups[userID] = append(userGroups[userID], groupID)
		}
	}

	for _, entry := range entries {
		pk := entry.pkInt()
		switch entry.Model {
		case "server.computer":
			computers = append(computers, computerOut{
				ID:           pk,
				Serial:       getString(entry.Fields, "serial"),
				Username:     getString(entry.Fields, "username"),
				ComputerName: getString(entry.Fields, "computername"),
				LastCheckin:  getString(entry.Fields, "last_checkin"),
			})
		case "server.secret":
			ciphertext := getString(entry.Fields, "secret")
			plaintext, err := decryptLegacySecret(ciphertext, legacyKey)
			if err != nil {
				return nil, fmt.Errorf("decrypt secret %d: %w", pk, err)
			}
			encrypted, err := newCodec.Encrypt(plaintext)
			if err != nil {
				return nil, fmt.Errorf("encrypt secret %d: %w", pk, err)
			}
			secrets = append(secrets, secretOut{
				ID:               pk,
				ComputerID:       getInt(entry.Fields, "computer"),
				SecretType:       getString(entry.Fields, "secret_type"),
				Secret:           encrypted,
				DateEscrowed:     getString(entry.Fields, "date_escrowed"),
				RotationRequired: getBool(entry.Fields, "rotation_required"),
			})
		case "server.request":
			requestingUser := usernameForID(usernames, getOptionalInt(entry.Fields, "requesting_user"))
			authUser := usernameForID(usernames, getOptionalInt(entry.Fields, "auth_user"))
			requests = append(requests, requestOut{
				ID:                pk,
				SecretID:          getInt(entry.Fields, "secret"),
				RequestingUser:    requestingUser,
				Approved:          getOptionalBool(entry.Fields, "approved"),
				AuthUser:          authUser,
				ReasonForRequest:  getString(entry.Fields, "reason_for_request"),
				ReasonForApproval: getString(entry.Fields, "reason_for_approval"),
				DateRequested:     getString(entry.Fields, "date_requested"),
				DateApproved:      getString(entry.Fields, "date_approved"),
				Current:           getBool(entry.Fields, "current"),
			})
		}
	}

	userList := make([]userOut, 0, len(users))
	for id, user := range users {
		groupNames := mapGroups(userGroups[id], groups)
		user.Groups = groupNames
		user.CanApprove = resolveCanApprove(user, userPermissions[id], groupPermissions, userGroups[id], permissionIDs)
		applyPasswordMapping(&user, passwordMap, emails[id])
		userList = append(userList, user)
	}
	sort.Slice(userList, func(i, j int) bool {
		return userList[i].ID < userList[j].ID
	})

	return &migrationOutput{
		Computers: computers,
		Secrets:   secrets,
		Requests:  requests,
		Users:     userList,
	}, nil
}

func mapGroups(groupIDs []int, groups map[int]string) []string {
	unique := map[string]struct{}{}
	for _, groupID := range groupIDs {
		if name := groups[groupID]; name != "" {
			unique[name] = struct{}{}
		}
	}
	names := make([]string, 0, len(unique))
	for name := range unique {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func resolveCanApprove(user userOut, userPerms []int, groupPerms map[int][]int, userGroupIDs []int, canApproveIDs map[int]struct{}) bool {
	if user.IsSuper {
		return true
	}
	if hasPermission(userPerms, canApproveIDs) {
		return true
	}
	for _, groupID := range userGroupIDs {
		if hasPermission(groupPerms[groupID], canApproveIDs) {
			return true
		}
	}
	return false
}

func hasPermission(permissionIDs []int, allowed map[int]struct{}) bool {
	for _, permID := range permissionIDs {
		if _, ok := allowed[permID]; ok {
			return true
		}
	}
	return false
}

func applyPasswordMapping(user *userOut, passwordMap map[string]passwordMapEntry, email string) {
	user.MustResetPassword = true
	user.LocalLoginEnabled = false
	user.AuthSource = "saml"
	if entry, ok := passwordMap[strings.ToLower(user.Username)]; ok {
		user.PasswordHash = entry.PasswordHash
		user.MustResetPassword = entry.MustResetPassword
		user.LocalLoginEnabled = true
		user.AuthSource = "local"
		return
	}
	if email != "" {
		if entry, ok := passwordMap[email]; ok {
			user.PasswordHash = entry.PasswordHash
			user.MustResetPassword = entry.MustResetPassword
			user.LocalLoginEnabled = true
			user.AuthSource = "local"
		}
	}
}

func decryptLegacySecret(value string, key *fernet.Key) (string, error) {
	if value == "" {
		return "", errors.New("empty secret")
	}
	// Try Fernet decryption first (for encrypted Django databases)
	plaintext := fernet.VerifyAndDecrypt([]byte(value), 0*time.Second, []*fernet.Key{key})
	if plaintext != nil {
		return string(plaintext), nil
	}
	// Fallback: if Fernet decryption fails, assume the secret is plaintext
	// (some Django installations may not have field encryption enabled)
	return value, nil
}

func marshalOutput(output *migrationOutput) ([]byte, error) {
	return json.MarshalIndent(output, "", "  ")
}

func usernameForID(users map[int]string, id *int) string {
	if id == nil {
		return ""
	}
	if username, ok := users[*id]; ok {
		return username
	}
	return fmt.Sprintf("user-%d", *id)
}

func getString(fields map[string]interface{}, key string) string {
	value, ok := fields[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func getInt(fields map[string]interface{}, key string) int {
	value, ok := fields[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case json.Number:
		parsed, _ := typed.Int64()
		return int(parsed)
	default:
		return 0
	}
}

func getOptionalInt(fields map[string]interface{}, key string) *int {
	value, ok := fields[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case float64:
		value := int(typed)
		return &value
	case int:
		return &typed
	case json.Number:
		parsed, _ := typed.Int64()
		value := int(parsed)
		return &value
	default:
		return nil
	}
}

func getBool(fields map[string]interface{}, key string) bool {
	value, ok := fields[key]
	if !ok || value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return false
	}
}

func getOptionalBool(fields map[string]interface{}, key string) *bool {
	value, ok := fields[key]
	if !ok {
		return nil
	}
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case bool:
		return &typed
	default:
		return nil
	}
}
