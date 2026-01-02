package donation

import (
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// sensitivePatterns defines patterns for sensitive information that should be filtered.
var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(admin_token|admintoken)["\s:=]+["\s]*([^"\s,}]+)`),
	regexp.MustCompile(`(?i)(client_secret|clientsecret)["\s:=]+["\s]*([^"\s,}]+)`),
	regexp.MustCompile(`(?i)(access_token|accesstoken)["\s:=]+["\s]*([^"\s,}]+)`),
	regexp.MustCompile(`(?i)(refresh_token|refreshtoken)["\s:=]+["\s]*([^"\s,}]+)`),
	regexp.MustCompile(`(?i)(api_key|apikey)["\s:=]+["\s]*([^"\s,}]+)`),
	regexp.MustCompile(`(?i)(password)["\s:=]+["\s]*([^"\s,}]+)`),
	regexp.MustCompile(`(?i)(secret)["\s:=]+["\s]*([^"\s,}]+)`),
	regexp.MustCompile(`(?i)(bearer\s+)([^\s"]+)`),
}

// sensitiveKeys defines keys that should be redacted in log fields.
var sensitiveKeys = map[string]bool{
	"admin_token":   true,
	"client_secret": true,
	"access_token":  true,
	"refresh_token": true,
	"api_key":       true,
	"password":      true,
	"secret":        true,
	"token":         true,
}

// DonationLogger provides structured logging for donation operations.
type DonationLogger struct {
	logger *log.Entry
}

// NewDonationLogger creates a new donation logger.
func NewDonationLogger() *DonationLogger {
	return &DonationLogger{
		logger: log.WithField("component", "donation"),
	}
}

// LogDonation logs a successful donation.
func (l *DonationLogger) LogDonation(linuxDoID, newAPIUserID int, quotaAmount int64) {
	l.logger.WithFields(log.Fields{
		"event":          "donation_success",
		"linux_do_id":    linuxDoID,
		"newapi_user_id": newAPIUserID,
		"quota_amount":   quotaAmount,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}).Info("donation processed successfully")
}

// LogDonationError logs a failed donation.
func (l *DonationLogger) LogDonationError(linuxDoID, newAPIUserID int, quotaAmount int64, err error) {
	l.logger.WithFields(log.Fields{
		"event":          "donation_failed",
		"linux_do_id":    linuxDoID,
		"newapi_user_id": newAPIUserID,
		"quota_amount":   quotaAmount,
		"error":          filterSensitive(err.Error()),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}).Error("donation failed")
}

// LogBinding logs a user binding event.
func (l *DonationLogger) LogBinding(linuxDoID, newAPIUserID int) {
	l.logger.WithFields(log.Fields{
		"event":          "user_binding",
		"linux_do_id":    linuxDoID,
		"newapi_user_id": newAPIUserID,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}).Info("user binding created")
}

// LogLogin logs a user login event.
func (l *DonationLogger) LogLogin(linuxDoID int, username, role string) {
	l.logger.WithFields(log.Fields{
		"event":       "user_login",
		"linux_do_id": linuxDoID,
		"username":    username,
		"role":        role,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}).Info("user logged in")
}

// LogLogout logs a user logout event.
func (l *DonationLogger) LogLogout(linuxDoID int) {
	l.logger.WithFields(log.Fields{
		"event":       "user_logout",
		"linux_do_id": linuxDoID,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}).Info("user logged out")
}

// LogError logs an error with context.
func (l *DonationLogger) LogError(operation string, err error, context map[string]interface{}) {
	fields := log.Fields{
		"event":     "error",
		"operation": operation,
		"error":     filterSensitive(err.Error()),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Add context fields, filtering sensitive data
	for k, v := range context {
		if sensitiveKeys[strings.ToLower(k)] {
			fields[k] = "[REDACTED]"
		} else if str, ok := v.(string); ok {
			fields[k] = filterSensitive(str)
		} else {
			fields[k] = v
		}
	}

	l.logger.WithFields(fields).Error("operation failed")
}

// LogOAuthError logs an OAuth-related error.
func (l *DonationLogger) LogOAuthError(stage string, err error) {
	l.logger.WithFields(log.Fields{
		"event":     "oauth_error",
		"stage":     stage,
		"error":     filterSensitive(err.Error()),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}).Error("OAuth error")
}

// LogNewAPIError logs a New-API related error.
func (l *DonationLogger) LogNewAPIError(operation string, userID int, err error) {
	l.logger.WithFields(log.Fields{
		"event":     "newapi_error",
		"operation": operation,
		"user_id":   userID,
		"error":     filterSensitive(err.Error()),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}).Error("New-API error")
}

// filterSensitive removes sensitive information from a string.
func filterSensitive(s string) string {
	result := s
	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// Keep the key name but redact the value
			parts := pattern.FindStringSubmatch(match)
			if len(parts) >= 2 {
				return strings.Replace(match, parts[len(parts)-1], "[REDACTED]", 1)
			}
			return "[REDACTED]"
		})
	}
	return result
}

// FilterSensitiveFields filters sensitive fields from a map.
func FilterSensitiveFields(fields map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		if sensitiveKeys[strings.ToLower(k)] {
			result[k] = "[REDACTED]"
		} else if str, ok := v.(string); ok {
			result[k] = filterSensitive(str)
		} else {
			result[k] = v
		}
	}
	return result
}
