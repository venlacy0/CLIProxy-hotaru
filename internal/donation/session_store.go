package donation

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const (
	// SessionTTL is the default session expiration time (24 hours).
	SessionTTL = 24 * time.Hour
	// TokenLength is the length of generated session tokens in bytes.
	TokenLength = 32
)

// SessionStore manages user sessions in memory.
type SessionStore struct {
	sessions sync.Map
	ttl      time.Duration
}

// NewSessionStore creates a new session store with the default TTL.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		ttl: SessionTTL,
	}
}

// NewSessionStoreWithTTL creates a new session store with a custom TTL.
func NewSessionStoreWithTTL(ttl time.Duration) *SessionStore {
	if ttl <= 0 {
		ttl = SessionTTL
	}
	return &SessionStore{
		ttl: ttl,
	}
}

// GenerateToken generates a cryptographically secure random token.
// Returns a hex-encoded string of TokenLength bytes.
func GenerateToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create creates a new session for the given user and role.
// Returns the created session with a generated token.
func (s *SessionStore) Create(user *LinuxDoUser, role string) (*Session, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:        token,
		LinuxDoID: user.ID,
		Username:  user.Username,
		Role:      role,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}

	s.sessions.Store(token, session)
	return session, nil
}

// Get retrieves a session by its token.
// Returns nil if the session doesn't exist or has expired.
func (s *SessionStore) Get(sessionID string) *Session {
	value, ok := s.sessions.Load(sessionID)
	if !ok {
		return nil
	}

	session, ok := value.(*Session)
	if !ok {
		return nil
	}

	// Check if session has expired
	if session.IsExpired() {
		s.sessions.Delete(sessionID)
		return nil
	}

	return session
}

// Update updates an existing session.
// Returns false if the session doesn't exist.
func (s *SessionStore) Update(session *Session) bool {
	if session == nil || session.ID == "" {
		return false
	}

	_, exists := s.sessions.Load(session.ID)
	if !exists {
		return false
	}

	s.sessions.Store(session.ID, session)
	return true
}

// Delete removes a session by its token.
func (s *SessionStore) Delete(sessionID string) {
	s.sessions.Delete(sessionID)
}

// Cleanup removes all expired sessions.
// This can be called periodically to free memory.
func (s *SessionStore) Cleanup() {
	now := time.Now()
	s.sessions.Range(func(key, value interface{}) bool {
		if session, ok := value.(*Session); ok {
			if now.After(session.ExpiresAt) {
				s.sessions.Delete(key)
			}
		}
		return true
	})
}
