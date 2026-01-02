package donation

import (
	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	log "github.com/sirupsen/logrus"
)

// DonationModule holds all donation-related services and handlers.
type DonationModule struct {
	handler       *DonationHandler
	sessionStore  *SessionStore
	bindingStore  *BindingStore
	roleService   *RoleService
	linuxDoSvc    *LinuxDoConnectService
	newAPISvc     *NewAPIService
	logger        *DonationLogger
	isConfigured  bool
}

// NewDonationModule creates a new donation module with all dependencies.
func NewDonationModule(cfg *config.Config) (*DonationModule, error) {
	// Check if Linux Do Connect is configured
	if cfg.LinuxDoConnect.ClientID == "" || cfg.LinuxDoConnect.ClientSecret == "" {
		log.Warn("Linux Do Connect is not configured, donation features will be disabled")
		return &DonationModule{isConfigured: false}, nil
	}

	// Initialize services
	sessionStore := NewSessionStore()
	
	bindingStore, err := NewBindingStore(cfg.AuthDir)
	if err != nil {
		return nil, err
	}

	roleService := NewRoleService(cfg.Donation.AdminLinuxDoIDs)
	linuxDoSvc := NewLinuxDoConnectService(cfg.LinuxDoConnect)
	newAPISvc := NewNewAPIService()
	logger := NewDonationLogger()

	quotaAmount := cfg.Donation.QuotaAmount
	if quotaAmount <= 0 {
		quotaAmount = 2000000 // Default $20
	}

	handler := NewDonationHandler(
		linuxDoSvc,
		newAPISvc,
		sessionStore,
		bindingStore,
		roleService,
		logger,
		quotaAmount,
	)

	return &DonationModule{
		handler:      handler,
		sessionStore: sessionStore,
		bindingStore: bindingStore,
		roleService:  roleService,
		linuxDoSvc:   linuxDoSvc,
		newAPISvc:    newAPISvc,
		logger:       logger,
		isConfigured: true,
	}, nil
}

// IsConfigured returns true if the donation module is properly configured.
func (m *DonationModule) IsConfigured() bool {
	return m != nil && m.isConfigured
}

// RegisterRoutes registers all donation-related routes on the given engine.
func (m *DonationModule) RegisterRoutes(engine *gin.Engine) {
	if !m.IsConfigured() {
		log.Info("Donation module not configured, skipping route registration")
		return
	}

	// Public routes (no auth required)
	linuxdo := engine.Group("/linuxdo")
	{
		linuxdo.GET("/login", m.handler.HandleLogin)
		linuxdo.GET("/callback", m.handler.HandleCallback)
	}

	// Protected routes (auth required)
	authMiddleware := AuthMiddleware(m.sessionStore)
	
	// Bind routes
	bind := engine.Group("/bind")
	bind.Use(authMiddleware)
	{
		bind.GET("", m.handler.HandleBindPage)
		bind.POST("", m.handler.HandleBind)
	}

	// Donate routes
	donate := engine.Group("/donate")
	donate.Use(authMiddleware)
	{
		donate.GET("", m.handler.HandleDonatePage)
		donate.POST("/confirm", m.handler.HandleDonateConfirm)
	}

	// Logout route
	engine.POST("/logout", authMiddleware, m.handler.HandleLogout)

	// Status route (optional auth - shows login status)
	engine.GET("/status", m.optionalAuthMiddleware(), m.handler.HandleStatus)

	// Index page (main HTML page)
	engine.GET("/", m.handler.HandleIndexPage)

	log.Info("Donation routes registered")
}

// optionalAuthMiddleware creates a middleware that optionally validates session.
// If session exists and is valid, it's stored in context. Otherwise, request continues.
func (m *DonationModule) optionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(SessionCookieName)
		if err != nil || sessionID == "" {
			c.Next()
			return
		}

		session := m.sessionStore.Get(sessionID)
		if session != nil {
			c.Set(SessionContextKey, session)
		}
		c.Next()
	}
}

// GetSessionStore returns the session store for use in other middleware.
func (m *DonationModule) GetSessionStore() *SessionStore {
	return m.sessionStore
}

// GetRoleService returns the role service for use in access control.
func (m *DonationModule) GetRoleService() *RoleService {
	return m.roleService
}

// IsAdmin checks if the current request is from an admin user.
func (m *DonationModule) IsAdmin(c *gin.Context) bool {
	session := GetSessionFromContext(c)
	if session == nil {
		return false
	}
	return session.Role == RoleAdmin
}

// CheckAdminAccess is a helper function to check admin access for auth file operations.
// Returns true if access should be allowed, false if it should be denied.
func (m *DonationModule) CheckAdminAccess(c *gin.Context) bool {
	if !m.IsConfigured() {
		// If donation module is not configured, allow all access (original behavior)
		return true
	}

	session := GetSessionFromContext(c)
	if session == nil {
		// No session means not logged in via donation site
		// Allow access (original behavior for management API)
		return true
	}

	// If logged in via donation site, check role
	return session.Role == RoleAdmin
}
