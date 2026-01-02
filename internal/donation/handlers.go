package donation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// DonationHandler handles HTTP requests for the donation site.
type DonationHandler struct {
	linuxDoService *LinuxDoConnectService
	newAPIService  *NewAPIService
	sessionStore   *SessionStore
	bindingStore   *BindingStore
	roleService    *RoleService
	logger         *DonationLogger
	quotaAmount    int64
}

// NewDonationHandler creates a new donation handler with all dependencies.
func NewDonationHandler(
	linuxDoService *LinuxDoConnectService,
	newAPIService *NewAPIService,
	sessionStore *SessionStore,
	bindingStore *BindingStore,
	roleService *RoleService,
	logger *DonationLogger,
	quotaAmount int64,
) *DonationHandler {
	if quotaAmount <= 0 {
		quotaAmount = 2000000 // Default $20
	}
	return &DonationHandler{
		linuxDoService: linuxDoService,
		newAPIService:  newAPIService,
		sessionStore:   sessionStore,
		bindingStore:   bindingStore,
		roleService:    roleService,
		logger:         logger,
		quotaAmount:    quotaAmount,
	}
}

// generateState generates a random state string for OAuth.
func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HandleLogin handles GET /linuxdo/login - redirects to Linux Do Connect OAuth.
func (h *DonationHandler) HandleLogin(c *gin.Context) {
	state, err := generateState()
	if err != nil {
		log.WithError(err).Error("failed to generate OAuth state")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to generate state",
		})
		return
	}

	// Store state in cookie for verification
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	authURL, err := h.linuxDoService.GenerateAuthURL(state)
	if err != nil {
		log.WithError(err).Error("failed to generate auth URL")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to generate authorization URL",
		})
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// HandleCallback handles GET /linuxdo/callback - OAuth callback from Linux Do Connect.
func (h *DonationHandler) HandleCallback(c *gin.Context) {
	ctx := context.Background()

	// Check for OAuth error
	if errStr := c.Query("error"); errStr != "" {
		errDesc := c.Query("error_description")
		log.WithFields(log.Fields{
			"error":       errStr,
			"description": errDesc,
		}).Warn("OAuth authorization denied")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errStr,
			"message": errDesc,
		})
		return
	}

	// Verify state
	state := c.Query("state")
	savedState, _ := c.Cookie("oauth_state")
	if state == "" || state != savedState {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_state",
			"message": "state mismatch",
		})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_code",
			"message": "authorization code not provided",
		})
		return
	}

	// Exchange code for token
	tokenResp, err := h.linuxDoService.ExchangeToken(ctx, code)
	if err != nil {
		log.WithError(err).Error("failed to exchange token")
		h.logger.LogError("token_exchange", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_exchange_failed",
			"message": "failed to exchange authorization code",
		})
		return
	}

	// Get user info
	user, err := h.linuxDoService.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		log.WithError(err).Error("failed to get user info")
		h.logger.LogError("get_user_info", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "user_info_failed",
			"message": "failed to get user information",
		})
		return
	}

	// Determine role
	role := h.roleService.DetermineRole(user.ID)

	// Create session
	session, err := h.sessionStore.Create(user, role)
	if err != nil {
		log.WithError(err).Error("failed to create session")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "session_failed",
			"message": "failed to create session",
		})
		return
	}

	// Check if user has binding
	binding := h.bindingStore.GetByLinuxDoID(user.ID)
	if binding != nil {
		session.NewAPIUserID = binding.NewAPIUserID
		h.sessionStore.Update(session)
	}

	// Set session cookie (24 hours)
	SetSessionCookie(c, session.ID, int(SessionTTL.Seconds()))

	log.WithFields(log.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     role,
	}).Info("user logged in")

	// Redirect to home or bind page
	if binding == nil {
		c.Redirect(http.StatusFound, "/bind")
	} else {
		c.Redirect(http.StatusFound, "/donate")
	}
}

// HandleLogout handles POST /logout - destroys session and clears cookie.
func (h *DonationHandler) HandleLogout(c *gin.Context) {
	session := GetSessionFromContext(c)
	if session != nil {
		h.sessionStore.Delete(session.ID)
		log.WithField("user_id", session.LinuxDoID).Info("user logged out")
	}

	ClearSessionCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "logged out successfully",
	})
}

// HandleBindPage handles GET /bind - returns binding status.
func (h *DonationHandler) HandleBindPage(c *gin.Context) {
	session := GetSessionFromContext(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	binding := h.bindingStore.GetByLinuxDoID(session.LinuxDoID)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"linux_do_id": session.LinuxDoID,
			"username":    session.Username,
			"role":        session.Role,
		},
		"bound":          binding != nil,
		"newapi_user_id": session.NewAPIUserID,
	})
}

// HandleBind handles POST /bind - binds a new-api user ID.
func (h *DonationHandler) HandleBind(c *gin.Context) {
	session := GetSessionFromContext(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Check if already bound
	if session.NewAPIUserID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "already_bound",
			"message": "user is already bound to a new-api account",
		})
		return
	}

	// Parse request
	var req struct {
		NewAPIUserID int `json:"newapi_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "newapi_user_id is required",
		})
		return
	}

	ctx := context.Background()

	// Get user from new-api
	newAPIUser, err := h.newAPIService.GetUserByID(ctx, req.NewAPIUserID)
	if err != nil {
		log.WithError(err).WithField("newapi_user_id", req.NewAPIUserID).Warn("failed to get new-api user")
		h.logger.LogError("get_newapi_user", err, map[string]interface{}{
			"newapi_user_id": req.NewAPIUserID,
		})
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "new-api user not found",
		})
		return
	}

	// Verify Linux Do ID matches
	newAPILinuxDoID, _ := strconv.Atoi(newAPIUser.LinuxDoID)
	if newAPILinuxDoID != session.LinuxDoID {
		log.WithFields(log.Fields{
			"session_linux_do_id": session.LinuxDoID,
			"newapi_linux_do_id":  newAPIUser.LinuxDoID,
		}).Warn("Linux Do ID mismatch")
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "id_mismatch",
			"message": "Linux Do ID does not match",
		})
		return
	}

	// Create binding
	binding := &UserBinding{
		LinuxDoID:    session.LinuxDoID,
		NewAPIUserID: req.NewAPIUserID,
		BoundAt:      time.Now(),
	}
	if err := h.bindingStore.Create(binding); err != nil {
		log.WithError(err).Error("failed to create binding")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "binding_failed",
			"message": "failed to save binding",
		})
		return
	}

	// Update session
	session.NewAPIUserID = req.NewAPIUserID
	h.sessionStore.Update(session)

	h.logger.LogBinding(session.LinuxDoID, req.NewAPIUserID)

	log.WithFields(log.Fields{
		"linux_do_id":    session.LinuxDoID,
		"newapi_user_id": req.NewAPIUserID,
	}).Info("user binding created")

	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"message":        "binding successful",
		"newapi_user_id": req.NewAPIUserID,
	})
}

// HandleDonatePage handles GET /donate - returns donation status.
func (h *DonationHandler) HandleDonatePage(c *gin.Context) {
	session := GetSessionFromContext(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"linux_do_id":    session.LinuxDoID,
			"username":       session.Username,
			"role":           session.Role,
			"newapi_user_id": session.NewAPIUserID,
		},
		"bound":        session.NewAPIUserID != 0,
		"quota_amount": h.quotaAmount,
	})
}

// HandleDonateConfirm handles POST /donate/confirm - confirms donation and adds quota.
func (h *DonationHandler) HandleDonateConfirm(c *gin.Context) {
	session := GetSessionFromContext(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Check if user is bound
	if session.NewAPIUserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "not_bound",
			"message": "please bind your new-api account first",
		})
		return
	}

	ctx := context.Background()

	// Add quota
	err := h.newAPIService.AddQuota(ctx, session.NewAPIUserID, h.quotaAmount)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user_id":      session.LinuxDoID,
			"newapi_id":    session.NewAPIUserID,
			"quota_amount": h.quotaAmount,
		}).Error("failed to add quota")
		h.logger.LogDonationError(session.LinuxDoID, session.NewAPIUserID, h.quotaAmount, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "quota_failed",
			"message": "failed to add quota, please contact administrator",
		})
		return
	}

	// Log successful donation
	h.logger.LogDonation(session.LinuxDoID, session.NewAPIUserID, h.quotaAmount)

	log.WithFields(log.Fields{
		"linux_do_id":    session.LinuxDoID,
		"newapi_user_id": session.NewAPIUserID,
		"quota_amount":   h.quotaAmount,
	}).Info("donation processed successfully")

	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"message":      "donation successful, quota added",
		"quota_amount": h.quotaAmount,
	})
}

// HandleStatus handles GET /status - returns current user status (public endpoint).
func (h *DonationHandler) HandleStatus(c *gin.Context) {
	session := GetSessionFromContext(c)
	if session == nil {
		c.JSON(http.StatusOK, gin.H{
			"logged_in": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logged_in": true,
		"user": gin.H{
			"linux_do_id":    session.LinuxDoID,
			"username":       session.Username,
			"role":           session.Role,
			"newapi_user_id": session.NewAPIUserID,
		},
		"bound": session.NewAPIUserID != 0,
	})
}

// HandleIndexPage handles GET / - serves the main HTML page.
func (h *DonationHandler) HandleIndexPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, GetIndexPageHTML())
}
