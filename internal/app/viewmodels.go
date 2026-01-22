package app

import "crypt-server/internal/store"

type User struct {
	ID                int
	Username          string
	IsAuthenticated   bool
	IsStaff           bool
	CanApprove        bool
	LocalLoginEnabled bool
	MustResetPassword bool
	AuthSource        string
}

type SecretView struct {
	Secret   *store.Secret
	Approved bool
	Pending  bool
}

type SecretChar struct {
	Char  string
	Class string
}

type RequestView struct {
	ID               int
	Serial           string
	ComputerName     string
	RequestingUser   string
	ReasonForRequest string
	DateRequested    string
}

type TemplateData struct {
	Title                         string
	User                          User
	Version                       string
	OutstandingCount              int
	Computers                     []*store.Computer
	Computer                      *store.Computer
	Secrets                       []*store.Secret
	SecretViews                   []SecretView
	Secret                        *store.Secret
	Requests                      []*store.Request
	ManageRequests                []RequestView
	Request                       *store.Request
	ErrorMessage                  string
	CanRequest                    bool
	RequestApproved               bool
	ApprovedRequestID             int
	RequestsForSecret             []*store.Request
	SecretChars                   []SecretChar
	Users                         []*store.User
	NewUser                       UserForm
	AdminUser                     *store.User
	AuditEvents                   []*store.AuditEvent
	AuditSearch                   string
	AuditPage                     int
	AuditPageSize                 int
	AuditTotal                    int
	AuditTotalPages               int
	CSRFToken                     string
	PasswordChangeRequiresCurrent bool
	SAMLAvailable                 bool
	SAMLLoginURL                  string
}

type UserForm struct {
	Username          string
	IsStaff           bool
	CanApprove        bool
	LocalLoginEnabled bool
	MustResetPassword bool
	AuthSource        string
}
