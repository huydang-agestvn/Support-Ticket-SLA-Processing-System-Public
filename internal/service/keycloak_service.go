package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"support-ticket.com/internal/dto"
)

type ClientRequest struct {
	tokenURL     string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewClient(tokenURL, clientID, clientSecret string) *ClientRequest {
	return &ClientRequest{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ClientRequest) Login(username, password string) (*dto.KeycloakTokenResponse, error) {
	if c.tokenURL == "" {
		return nil, fmt.Errorf("keycloak token url is required")
	}

	if c.clientID == "" {
		return nil, fmt.Errorf("keycloak client id is required")
	}

	if c.clientSecret == "" {
		return nil, fmt.Errorf("keycloak client secret is required")
	}

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequest(
		http.MethodPost,
		c.tokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create keycloak login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call keycloak token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var kcErr dto.KeycloakErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&kcErr)

		if kcErr.Error != "" {
			return nil, fmt.Errorf("keycloak login failed: %s - %s", kcErr.Error, kcErr.ErrorDescription)
		}

		return nil, fmt.Errorf("keycloak login failed with status code: %d", resp.StatusCode)
	}

	var tokenResp dto.KeycloakTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode keycloak token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("keycloak response missing access_token")
	}

	return &tokenResp, nil
}
