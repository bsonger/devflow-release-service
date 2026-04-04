package router

import (
	"github.com/bsonger/devflow-release-service/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterReleaseRoutes(rg *gin.RouterGroup) {
	release := rg.Group("/releases")

	release.GET("", api.ReleaseRouteApi.List)
	release.GET("/:id", api.ReleaseRouteApi.Get)
	release.POST("", api.ReleaseRouteApi.Create)
	//release.PUT("/:id", api.ReleaseRouteApi.Update)
	//release.DELETE("/:id", api.ReleaseRouteApi.Delete)
}
