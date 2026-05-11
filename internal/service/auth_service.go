package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AuthService struct {
	tokenURL     string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResult struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
}

type keycloakTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
}

type keycloakErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func NewAuthService(tokenURL, clientID, clientSecret string) *AuthService {
	return &AuthService{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *AuthService) Login(input LoginInput) (*LoginResult, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)

	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", s.clientID)
	form.Set("client_secret", s.clientSecret)
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequest(
		http.MethodPost,
		s.tokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call keycloak token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var kcErr keycloakErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&kcErr)

		if kcErr.Error != "" {
			return nil, fmt.Errorf("keycloak login failed: %s - %s", kcErr.Error, kcErr.ErrorDescription)
		}

		return nil, fmt.Errorf("keycloak login failed with status code: %d", resp.StatusCode)
	}

	var tokenResp keycloakTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode keycloak token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("keycloak response missing access_token")
	}

	return &LoginResult{
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		TokenType:        tokenResp.TokenType,
		ExpiresIn:        tokenResp.ExpiresIn,
		RefreshExpiresIn: tokenResp.RefreshExpiresIn,
		Scope:            tokenResp.Scope,
	}, nil
}
