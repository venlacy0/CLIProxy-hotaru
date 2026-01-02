// Package donation provides donation site functionality including Linux Do Connect OAuth,
// user binding, and donation processing.
package donation

import "time"

// LinuxDoUser represents a user from Linux Do Connect OAuth.
type LinuxDoUser struct {
	// ID is the unique identifier of the user on Linux Do platform.
	ID int `json:"id"`
	// Username is the user's login name.
	Username string `json:"username"`
	// Name is the user's display name.
	Name string `json:"name"`
	// Email is the user's email address.
	Email string `json:"email"`
	// Avatar is the URL to the user's avatar image.
	Avatar string `json:"avatar_url"`
}

// Session represents a user session for the donation site.
type Session struct {
	// ID is the unique session identifier (token).
	ID string `json:"id"`
	// LinuxDoID is the user's Linux Do platform ID.
	LinuxDoID int `json:"linux_do_id"`
	// Username is the user's login name from Linux Do.
	Username string `json:"username"`
	// Role is the user's role, either "user" or "admin".
	Role string `json:"role"`
	// NewAPIUserID is the bound new-api user ID (0 if not bound).
	NewAPIUserID int `json:"newapi_user_id,omitempty"`
	// CreatedAt is when the session was created.
	CreatedAt time.Time `json:"created_at"`
	// ExpiresAt is when the session expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// UserBinding represents the binding relationship between a Linux Do user and a new-api user.
type UserBinding struct {
	// LinuxDoID is the user's Linux Do platform ID.
	LinuxDoID int `json:"linux_do_id"`
	// NewAPIUserID is the user's new-api platform ID.
	NewAPIUserID int `json:"newapi_user_id"`
	// BoundAt is when the binding was created.
	BoundAt time.Time `json:"bound_at"`
}

// NewAPIUser represents a user from the new-api platform.
type NewAPIUser struct {
	// ID is the user's ID on new-api platform.
	ID int `json:"id"`
	// Username is the user's login name.
	Username string `json:"username"`
	// LinuxDoID is the user's Linux Do platform ID (for verification).
	LinuxDoID string `json:"linux_do_id"`
	// Quota is the user's current quota balance.
	Quota int64 `json:"quota"`
}
