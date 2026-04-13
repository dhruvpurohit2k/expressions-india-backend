package auth

import (
	"net/http"
	"strings"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

const claimsKey = "auth_claims"

// extractToken reads the JWT from the Authorization: Bearer header first,
// then falls back to the ei_access cookie. This lets mobile clients use Bearer
// tokens while web clients continue to use HTTP-only cookies.
func extractToken(c *gin.Context) string {
	if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	token, _ := c.Cookie(cookieAccess)
	return token
}

// RequireAdmin blocks requests that don't carry a valid admin JWT (cookie or Bearer).
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			utils.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing or malformed token")
			c.Abort()
			return
		}
		claims, err := ParseAccessToken(tokenStr)
		if err != nil {
			utils.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			c.Abort()
			return
		}
		if !claims.IsAdmin {
			utils.Fail(c, http.StatusForbidden, "FORBIDDEN", "admin access required")
			c.Abort()
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// RequireAuth blocks requests without a valid JWT (cookie or Bearer) from any user.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			utils.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "login required")
			c.Abort()
			return
		}
		claims, err := ParseAccessToken(tokenStr)
		if err != nil {
			utils.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			c.Abort()
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// TryExtractClaims optionally reads the JWT (cookie or Bearer) and stores claims in
// context if valid. Does not abort — callers check GetClaims themselves.
func TryExtractClaims() gin.HandlerFunc {
	return func(c *gin.Context) {
		if tokenStr := extractToken(c); tokenStr != "" {
			if claims, err := ParseAccessToken(tokenStr); err == nil {
				c.Set(claimsKey, claims)
			}
		}
		c.Next()
	}
}

// GetClaims retrieves auth claims stored by any auth middleware. Returns nil if not set.
func GetClaims(c *gin.Context) *Claims {
	v, _ := c.Get(claimsKey)
	claims, _ := v.(*Claims)
	return claims
}
