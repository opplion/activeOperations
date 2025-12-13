package router

import (
	"github.com/gin-gonic/gin"
	"activeOperations/internal/agent/middleware"
	"activeOperations/internal/agent/controller"
)

func StartServer() *gin.Engine {
	app := GetGinApp()
	app = middleware.HttpMetrics(app)
	AddRoutes(app)
	return app
}

func GetGinApp() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.Recovery())
	app.Any("/api/ping", func(c *gin.Context) { c.JSON(200, map[string]interface{}{"msg": "pong"}) })
	return app
}

func AddRoutes(app *gin.Engine) {
	app.POST("/api/chat/send",controller.StartChat)
	app.POST("/api/chat/cancel",controller.EndChat)
}
