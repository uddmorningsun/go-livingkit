package ginlib

import (
	"github.com/gin-gonic/gin"
)

// NewGinServer returns a gin.Engine instance with the series of middleware.
func NewGinServer() *gin.Engine {
	// gin.DisableConsoleColor()
	// gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	{
		engine.HandleMethodNotAllowed = true
		engine.RedirectTrailingSlash = false
		engine.Use(
			RecoverJSONResponse(nil),
		)
		engine.GET("/apis", APIs(engine))
	}
	return engine
}
