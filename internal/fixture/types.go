package fixture

// MigrationOutput is the structure of the converted fixture JSON.
type MigrationOutput struct {
	Computers []Computer `json:"computers"`
	Secrets   []Secret   `json:"secrets"`
	Requests  []Request  `json:"requests"`
	Users     []User     `json:"users"`
}

// Computer represents a computer entry from the fixture.
type Computer struct {
	ID           int    `json:"id"`
	Serial       string `json:"serial"`
	Username     string `json:"username"`
	ComputerName string `json:"computername"`
	LastCheckin  string `json:"last_checkin"`
}

// Secret represents a secret entry from the fixture.
// The Secret field is already encrypted with the new key.
type Secret struct {
	ID               int    `json:"id"`
	ComputerID       int    `json:"computer_id"`
	SecretType       string `json:"secret_type"`
	Secret           string `json:"secret"`
	DateEscrowed     string `json:"date_escrowed"`
	RotationRequired bool   `json:"rotation_required"`
}

// Request represents a request entry from the fixture.
type Request struct {
	ID                int    `json:"id"`
	SecretID          int    `json:"secret_id"`
	RequestingUser    string `json:"requesting_user"`
	Approved          *bool  `json:"approved"`
	AuthUser          string `json:"auth_user"`
	ReasonForRequest  string `json:"reason_for_request"`
	ReasonForApproval string `json:"reason_for_approval"`
	DateRequested     string `json:"date_requested"`
	DateApproved      string `json:"date_approved"`
	Current           bool   `json:"current"`
}

// User represents a user entry from the fixture.
type User struct {
	ID                int      `json:"id"`
	Username          string   `json:"username"`
	Email             string   `json:"email"`
	IsStaff           bool     `json:"is_staff"`
	IsSuper           bool     `json:"is_superuser"`
	CanApprove        bool     `json:"can_approve"`
	Groups            []string `json:"groups"`
	PasswordHash      string   `json:"password_hash"`
	MustResetPassword bool     `json:"must_reset_password"`
	LocalLoginEnabled bool     `json:"local_login_enabled"`
	AuthSource        string   `json:"auth_source"`
}
