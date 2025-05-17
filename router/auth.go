package router

import (
	"h-ui/controller"
	"h-ui/middleware"

	"github.com/gin-gonic/gin"
)

func initAuthRouter(authApi *gin.RouterGroup) {
	auth := authApi.Group("/auth")
	{
		// Only add the middleware if both domain and path are set
		allowedDomainSet := false
		securityPathSet := false
		if config, err := middleware.GetConfigValue("HUI_ALLOWED_DOMAIN"); err == nil && config != "" {
			allowedDomainSet = true
		}
		if config, err := middleware.GetConfigValue("HUI_SECURITY_PATH"); err == nil && config != "" {
			securityPathSet = true
		}
		if allowedDomainSet && securityPathSet {
			auth.POST("/login", middleware.DomainPathRestrictHandler(), controller.Login)
		} else {
			auth.POST("/login", controller.Login)
		}
	}
}
