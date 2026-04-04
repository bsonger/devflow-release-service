package router

import (
	"github.com/bsonger/devflow-release-service/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterIntentRoutes(rg *gin.RouterGroup) {
	intent := rg.Group("/intents")

	intent.GET("", api.IntentRouteApi.List)
	intent.GET("/:id", api.IntentRouteApi.Get)
}
