package donation

// RoleService handles role assignment for users.
type RoleService struct {
	adminIDs map[int]struct{}
}

// NewRoleService creates a new role service with the given admin user IDs.
func NewRoleService(adminLinuxDoIDs []int) *RoleService {
	adminIDs := make(map[int]struct{}, len(adminLinuxDoIDs))
	for _, id := range adminLinuxDoIDs {
		adminIDs[id] = struct{}{}
	}
	return &RoleService{
		adminIDs: adminIDs,
	}
}

// DetermineRole determines the role for a user based on their Linux Do ID.
// Returns "admin" if the user is in the admin list, otherwise "user".
func (s *RoleService) DetermineRole(linuxDoID int) string {
	if s == nil || s.adminIDs == nil {
		return RoleUser
	}
	if _, isAdmin := s.adminIDs[linuxDoID]; isAdmin {
		return RoleAdmin
	}
	return RoleUser
}

// IsAdmin returns true if the given Linux Do ID belongs to an admin.
func (s *RoleService) IsAdmin(linuxDoID int) bool {
	if s == nil || s.adminIDs == nil {
		return false
	}
	_, isAdmin := s.adminIDs[linuxDoID]
	return isAdmin
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
