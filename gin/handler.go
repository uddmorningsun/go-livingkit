package ginlib

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

var (
	// ginTreeNodeAPIPathRE will match `/api/:uuid` or `/api/:uuid/article/:id` or `/api/v1/*path`.
	ginTreeNodeAPIPathRE = regexp.MustCompile(`(^/.*)([:|*](\w+))(.*)`)
)

// APIs will response all registered API and format style from `/api/:uuid` to `/api/{uuid}`.
func APIs(engine *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		apis := make([]interface{}, 0)
		for _, value := range engine.Routes() {
			path := value.Path
			for ginTreeNodeAPIPathRE.MatchString(path) {
				path = ginTreeNodeAPIPathRE.ReplaceAllString(path, "$1{$3}$4")
			}
			apis = append(apis, map[string]interface{}{
				"path": path,
				"method": value.Method,
				"lastHandlerName": value.Handler,
			})
		}
		ResponseOK(c, http.StatusOK, map[string][]interface{}{
			"apis": apis,
		})
	}
}
