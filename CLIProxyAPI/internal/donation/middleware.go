package donation

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// SessionCookieName is the name of the session cookie.
	SessionCookieName = "session_id"
	// SessionContextKey is the key used to store session in gin.Context.
	SessionContextKey = "donation_session"
	// RoleUser is the role for regular users.
	RoleUser = "user"
	// RoleAdmin is the role for administrators.
	RoleAdmin = "admin"
)

// AuthMiddleware creates a middleware that validates session tokens.
// It reads the session_id from cookies and validates it against the session store.
// If valid, the session is stored in the context for downstream handlers.
func AuthMiddleware(sessionStore *SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session ID from cookie
		sessionID, err := c.Cookie(SessionCookieName)
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "session not found",
			})
			return
		}

		// Validate session
		session := sessionStore.Get(sessionID)
		if session == nil {
			// Clear invalid cookie
			c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid or expired session",
			})
			return
		}

		// Store session in context
		c.Set(SessionContextKey, session)
		c.Next()
	}
}

// RoleMiddleware creates a middleware that checks if the user has the required role.
// It must be used after AuthMiddleware.
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSessionFromContext(c)
		if session == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "session not found in context",
			})
			return
		}

		// Check role
		if !hasRequiredRole(session.Role, requiredRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// AdminOnlyMiddleware creates a middleware that only allows admin users.
func AdminOnlyMiddleware() gin.HandlerFunc {
	return RoleMiddleware(RoleAdmin)
}

// GetSessionFromContext retrieves the session from the gin context.
// Returns nil if no session is found.
func GetSessionFromContext(c *gin.Context) *Session {
	value, exists := c.Get(SessionContextKey)
	if !exists {
		return nil
	}
	session, ok := value.(*Session)
	if !ok {
		return nil
	}
	return session
}

// hasRequiredRole checks if the user's role meets the required role.
// Admin role has access to everything.
func hasRequiredRole(userRole, requiredRole string) bool {
	// Admin has access to everything
	if userRole == RoleAdmin {
		return true
	}
	// For user role, check exact match
	return userRole == requiredRole
}

// SetSessionCookie sets the session cookie in the response.
func SetSessionCookie(c *gin.Context, sessionID string, maxAge int) {
	c.SetCookie(
		SessionCookieName,
		sessionID,
		maxAge,
		"/",
		"",    // domain - empty means current domain
		false, // secure - set to true in production with HTTPS
		true,  // httpOnly - prevents JavaScript access
	)
}

// ClearSessionCookie clears the session cookie.
func ClearSessionCookie(c *gin.Context) {
	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
}

// IsAdmin returns true if the session belongs to an admin user.
func IsAdmin(c *gin.Context) bool {
	session := GetSessionFromContext(c)
	if session == nil {
		return false
	}
	return session.Role == RoleAdmin
}

// GetCurrentUserID returns the Linux Do user ID from the current session.
// Returns 0 if no session is found.
func GetCurrentUserID(c *gin.Context) int {
	session := GetSessionFromContext(c)
	if session == nil {
		return 0
	}
	return session.LinuxDoID
}
