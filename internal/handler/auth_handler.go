package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"support-ticket.com/internal/dto/common"
	"support-ticket.com/internal/dto/request"
	"support-ticket.com/internal/dto/response"
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
// @Param request body request.LoginRequest true "Login request"
// @Success 200 {object} map[string]interface{} "Login successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid username or password"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input request.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, common.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	result, err := h.authService.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, common.APIResponse[*response.LoginResponse]{
		Success: true,
		Data:    result,
	})
}
