package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AofArthit21/Hospital-information-system/internal/auth"
)

const (
	staffIDKey    = "staff_id"
	hospitalIDKey = "hospital_id"
)

func authMiddleware(tokens *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "a Bearer token is required")
			return
		}
		staffID, hospitalID, err := tokens.Parse(parts[1])
		if err != nil {
			writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "the access token is invalid or expired")
			return
		}
		c.Set(staffIDKey, staffID)
		c.Set(hospitalIDKey, hospitalID)
		c.Next()
	}
}

func hospitalIDFromContext(c *gin.Context) int64 {
	value, _ := c.Get(hospitalIDKey)
	hospitalID, _ := value.(int64)
	return hospitalID
}
