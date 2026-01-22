package app

import "time"

type Settings struct {
	ApproveOwn             bool
	AllApprove             bool
	SessionTTL             time.Duration
	CookieSecure           bool
	RequestCleanupInterval time.Duration
	RotateViewedSecrets    bool
}
