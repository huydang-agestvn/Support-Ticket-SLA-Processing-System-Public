package service

import (
	"fmt"
	"strings"

	"support-ticket.com/internal/dto/request"
	"support-ticket.com/internal/dto/response"
)

type AuthService struct {
	keycloakClient *ClientRequest
}

func NewAuthService(keycloakClient *ClientRequest) *AuthService {
	return &AuthService{
		keycloakClient: keycloakClient,
	}
}

func (s *AuthService) Login(input request.LoginRequest) (*response.LoginResponse, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)

	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	tokenResp, err := s.keycloakClient.Login(username, password)
	if err != nil {
		return nil, err
	}

	return &response.LoginResponse{
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		TokenType:        tokenResp.TokenType,
		ExpiresIn:        tokenResp.ExpiresIn,
		RefreshExpiresIn: tokenResp.RefreshExpiresIn,
		Scope:            tokenResp.Scope,
	}, nil
}
