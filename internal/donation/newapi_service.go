package donation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	// Environment variable names for New-API configuration
	EnvNewAPIBaseURL    = "NEW_API_BASE_URL"
	EnvNewAPIAdminToken = "NEW_API_ADMIN_TOKEN"
)

// NewAPIService handles interactions with the new-api platform.
type NewAPIService struct {
	baseURL    string
	adminToken string
	httpClient *http.Client
}

// NewNewAPIService creates a new New-API service.
// It reads configuration from environment variables.
func NewNewAPIService() *NewAPIService {
	return &NewAPIService{
		baseURL:    os.Getenv(EnvNewAPIBaseURL),
		adminToken: os.Getenv(EnvNewAPIAdminToken),
		httpClient: &http.Client{},
	}
}

// NewNewAPIServiceWithConfig creates a new New-API service with explicit configuration.
func NewNewAPIServiceWithConfig(baseURL, adminToken string) *NewAPIService {
	return &NewAPIService{
		baseURL:    baseURL,
		adminToken: adminToken,
		httpClient: &http.Client{},
	}
}

// SetHTTPClient sets a custom HTTP client (useful for testing or proxy support).
func (s *NewAPIService) SetHTTPClient(client *http.Client) {
	if client != nil {
		s.httpClient = client
	}
}

// IsConfigured returns true if the service has the minimum required configuration.
func (s *NewAPIService) IsConfigured() bool {
	return s.baseURL != "" && s.adminToken != ""
}

// GetUserByID retrieves a user by their ID from new-api.
func (s *NewAPIService) GetUserByID(ctx context.Context, userID int) (*NewAPIUser, error) {
	if !s.IsConfigured() {
		return nil, fmt.Errorf("new-api service is not configured")
	}

	url := fmt.Sprintf("%s/api/user/%d", s.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.setAuthHeader(req)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - new-api typically wraps data in a response object
	var response struct {
		Success bool       `json:"success"`
		Message string     `json:"message"`
		Data    NewAPIUser `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		// Try direct unmarshal if not wrapped
		var user NewAPIUser
		if err2 := json.Unmarshal(body, &user); err2 != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return &user, nil
	}

	if !response.Success {
		return nil, fmt.Errorf("get user failed: %s", response.Message)
	}

	return &response.Data, nil
}

// AddQuota adds quota to a user's account.
func (s *NewAPIService) AddQuota(ctx context.Context, userID int, amount int64) error {
	if !s.IsConfigured() {
		return fmt.Errorf("new-api service is not configured")
	}

	url := fmt.Sprintf("%s/api/user/quota", s.baseURL)
	payload := map[string]interface{}{
		"user_id": userID,
		"quota":   amount,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	s.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add quota: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add quota failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Check response for success
	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &response); err == nil {
		if !response.Success {
			return fmt.Errorf("add quota failed: %s", response.Message)
		}
	}

	return nil
}

// setAuthHeader sets the Authorization header with the admin token.
func (s *NewAPIService) setAuthHeader(req *http.Request) {
	if s.adminToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.adminToken)
	}
}

// GetBaseURL returns the configured base URL.
func (s *NewAPIService) GetBaseURL() string {
	return s.baseURL
}
