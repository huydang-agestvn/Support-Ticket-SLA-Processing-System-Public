package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"support-ticket.com/internal/auth"
)

type AuthMiddleware struct {
	authenticator *auth.KeycloakAuthenticator
}

func NewAuthMiddleware(authenticator *auth.KeycloakAuthenticator) *AuthMiddleware {
	return &AuthMiddleware{
		authenticator: authenticator,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing authorization header",
			})
			c.Abort()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing bearer token",
			})
			c.Abort()
			return
		}

		currentUser, err := m.authenticator.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid token: " + err.Error(),
			})
			c.Abort()
			return
		}

		ctx := auth.WithUser(c.Request.Context(), currentUser)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, ok := auth.UserFromContext(c.Request.Context())
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthenticated",
			})
			c.Abort()
			return
		}

		if !currentUser.HasAnyRole(allowedRoles...) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "forbidden: insufficient role",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
