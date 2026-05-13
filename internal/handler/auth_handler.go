package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"support-ticket.com/internal/dto"
	"support-ticket.com/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login godoc
// @Summary Login
// @Description Login with username and password through Keycloak
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} map[string]interface{} "Login successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid username or password"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	result, err := h.authService.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse[*dto.LoginResponse]{
		Success: true,
		Data:    result,
	})
}
