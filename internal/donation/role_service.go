package donation

import "strings"

// RoleService handles role assignment for users.
type RoleService struct {
	adminIDs       map[int]struct{}
	adminUsernames map[string]struct{}
}

// NewRoleService creates a new role service with the given admin user IDs and usernames.
func NewRoleService(adminLinuxDoIDs []int, adminUsernames []string) *RoleService {
	adminIDs := make(map[int]struct{}, len(adminLinuxDoIDs))
	for _, id := range adminLinuxDoIDs {
		adminIDs[id] = struct{}{}
	}
	adminUsernameMap := make(map[string]struct{}, len(adminUsernames))
	for _, username := range adminUsernames {
		// Store usernames in lowercase for case-insensitive matching
		adminUsernameMap[strings.ToLower(username)] = struct{}{}
	}
	return &RoleService{
		adminIDs:       adminIDs,
		adminUsernames: adminUsernameMap,
	}
}

// DetermineRole determines the role for a user based on their Linux Do ID or username.
// Returns "admin" if the user is in the admin list, otherwise "user".
func (s *RoleService) DetermineRole(linuxDoID int, username string) string {
	if s == nil {
		return RoleUser
	}
	// Check by ID first
	if s.adminIDs != nil {
		if _, isAdmin := s.adminIDs[linuxDoID]; isAdmin {
			return RoleAdmin
		}
	}
	// Check by username (case-insensitive)
	if s.adminUsernames != nil && username != "" {
		if _, isAdmin := s.adminUsernames[strings.ToLower(username)]; isAdmin {
			return RoleAdmin
		}
	}
	return RoleUser
}

// IsAdmin returns true if the given Linux Do ID or username belongs to an admin.
func (s *RoleService) IsAdmin(linuxDoID int, username string) bool {
	return s.DetermineRole(linuxDoID, username) == RoleAdmin
}

// AddAdmin adds a user ID to the admin list.
func (s *RoleService) AddAdmin(linuxDoID int) {
	if s.adminIDs == nil {
		s.adminIDs = make(map[int]struct{})
	}
	s.adminIDs[linuxDoID] = struct{}{}
}

// RemoveAdmin removes a user ID from the admin list.
func (s *RoleService) RemoveAdmin(linuxDoID int) {
	if s.adminIDs != nil {
		delete(s.adminIDs, linuxDoID)
	}
}

// GetAdminIDs returns a slice of all admin user IDs.
func (s *RoleService) GetAdminIDs() []int {
	if s == nil || s.adminIDs == nil {
		return nil
	}
	ids := make([]int, 0, len(s.adminIDs))
	for id := range s.adminIDs {
		ids = append(ids, id)
	}
	return ids
}
