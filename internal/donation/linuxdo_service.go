package donation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
)

const (
	// Default Linux Do Connect OAuth URLs
	DefaultAuthURL     = "https://connect.linux.do/oauth2/authorize"
	DefaultTokenURL    = "https://connect.linux.do/oauth2/token"
	DefaultUserInfoURL = "https://connect.linux.do/api/user"
)

// LinuxDoConnectService handles Linux Do Connect OAuth operations.
type LinuxDoConnectService struct {
	clientID     string
	clientSecret string
	redirectURI  string
	authURL      string
	tokenURL     string
	userInfoURL  string
	httpClient   *http.Client
}

// TokenResponse represents the OAuth token response from Linux Do Connect.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// NewLinuxDoConnectService creates a new Linux Do Connect service.
func NewLinuxDoConnectService(cfg config.LinuxDoConnectConfig) *LinuxDoConnectService {
	authURL := cfg.AuthURL
	if authURL == "" {
		authURL = DefaultAuthURL
	}
	tokenURL := cfg.TokenURL
	if tokenURL == "" {
		tokenURL = DefaultTokenURL
	}
	userInfoURL := cfg.UserInfoURL
	if userInfoURL == "" {
		userInfoURL = DefaultUserInfoURL
	}

	return &LinuxDoConnectService{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		redirectURI:  cfg.RedirectURI,
		authURL:      authURL,
		tokenURL:     tokenURL,
		userInfoURL:  userInfoURL,
		httpClient:   &http.Client{},
	}
}

// SetHTTPClient sets a custom HTTP client (useful for testing or proxy support).
func (s *LinuxDoConnectService) SetHTTPClient(client *http.Client) {
	if client != nil {
		s.httpClient = client
	}
}

// GenerateAuthURL generates the OAuth authorization URL.
// The state parameter should be a random string to prevent CSRF attacks.
func (s *LinuxDoConnectService) GenerateAuthURL(state string) (string, error) {
	if s.clientID == "" {
		return "", fmt.Errorf("client_id is not configured")
	}
	if s.redirectURI == "" {
		return "", fmt.Errorf("redirect_uri is not configured")
	}

	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("redirect_uri", s.redirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)

	return fmt.Sprintf("%s?%s", s.authURL, params.Encode()), nil
}

// ExchangeToken exchanges an authorization code for an access token.
func (s *LinuxDoConnectService) ExchangeToken(ctx context.Context, code string) (*TokenResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("authorization code is empty")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.redirectURI)
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", s.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo retrieves user information using the access token.
func (s *LinuxDoConnectService) GetUserInfo(ctx context.Context, accessToken string) (*LinuxDoUser, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is empty")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", s.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user info failed with status %d: %s", resp.StatusCode, string(body))
	}

	var user LinuxDoUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	return &user, nil
}

// IsConfigured returns true if the service has the minimum required configuration.
func (s *LinuxDoConnectService) IsConfigured() bool {
	return s.clientID != "" && s.clientSecret != "" && s.redirectURI != ""
}
