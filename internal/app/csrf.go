package app

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

type CSRFManager struct {
	cookieName string
	tokenBytes int
}

func NewCSRFManager(cookieName string, tokenBytes int) *CSRFManager {
	return &CSRFManager{cookieName: cookieName, tokenBytes: tokenBytes}
}

func (m *CSRFManager) EnsureToken(w http.ResponseWriter, r *http.Request, secure bool) (string, error) {
	if token := m.TokenFromRequest(r); token != "" {
		return token, nil
	}
	token, err := m.GenerateToken()
	if err != nil {
		return "", err
	}
	m.SetCookie(w, token, secure)
	return token, nil
}

func (m *CSRFManager) GenerateToken() (string, error) {
	buf := make([]byte, m.tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(buf), nil
}

func (m *CSRFManager) TokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (m *CSRFManager) SetCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func (m *CSRFManager) ValidateRequest(r *http.Request) bool {
	cookieToken := m.TokenFromRequest(r)
	if cookieToken == "" {
		return false
	}
	if err := r.ParseForm(); err != nil {
		return false
	}
	formToken := r.FormValue("csrf_token")
	if formToken == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookieToken), []byte(formToken)) == 1
}
