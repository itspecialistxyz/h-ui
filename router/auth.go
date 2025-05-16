package router

import (
	"h-ui/controller"
	"h-ui/middleware"

	"github.com/gin-gonic/gin"
)

func initAuthRouter(authApi *gin.RouterGroup) {
	auth := authApi.Group("/auth")
	{
		auth.POST("/login", middleware.DomainPathRestrictHandler(), controller.Login)
	}
}
