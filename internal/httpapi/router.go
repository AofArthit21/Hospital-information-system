package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/AofArthit21/Hospital-information-system/internal/auth"
)

func NewRouter(handler *Handler, tokens *auth.Manager) *gin.Engine {
	binding.EnableDecoderDisallowUnknownFields = true
	router := gin.New()
	router.Use(sanitizedLogger(), gin.Recovery(), requestSizeLimit(1<<20), securityHeaders())
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/staff/create", handler.createStaff)
	router.POST("/staff/login", handler.login)
	router.GET("/patient/search", authMiddleware(tokens), handler.searchPatients)
	return router
}

func sanitizedLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		path := strings.SplitN(params.Path, "?", 2)[0]
		return fmt.Sprintf("%s - [%s] \"%s %s\" %d %s\n",
			params.ClientIP,
			params.TimeStamp.Format(time.RFC3339),
			params.Method,
			path,
			params.StatusCode,
			params.Latency,
		)
	})
}

func requestSizeLimit(bytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, bytes)
		c.Next()
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	}
}
