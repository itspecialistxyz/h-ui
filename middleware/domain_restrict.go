package middleware

import (
	"net/http"
	"os"
	"strings"

	"h-ui/model/constant"
	"h-ui/service"

	"github.com/gin-gonic/gin"
)

// DomainPathRestrictHandler restricts access to sensitive endpoints by domain and path.
func DomainPathRestrictHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Fetch from config table (DB), fallback to env var
		allowedDomain := ""
		securityPath := ""
		if config, err := service.GetConfig(constant.HUIAllowedDomain); err == nil && config.Value != nil && *config.Value != "" {
			allowedDomain = *config.Value
		} else {
			allowedDomain = os.Getenv("HUI_ALLOWED_DOMAIN")
		}
		if config, err := service.GetConfig(constant.HUISecurityPath); err == nil && config.Value != nil && *config.Value != "" {
			securityPath = *config.Value
		} else {
			securityPath = os.Getenv("HUI_SECURITY_PATH")
		}

		if allowedDomain == "" || securityPath == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Server misconfiguration: domain or security path not set"})
			return
		}
		if c.Request.Host != allowedDomain || !strings.HasPrefix(c.Request.URL.Path, securityPath) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: domain or path not allowed"})
			return
		}
		c.Next()
	}
}
