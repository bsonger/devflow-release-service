package router

import (
	"github.com/bsonger/devflow-release-service/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterManifestRoutes(rg *gin.RouterGroup) {
	manifests := rg.Group("/manifests")
	manifests.POST("", api.ManifestRouteApi.Create)
	manifests.GET("", api.ManifestRouteApi.List)
	manifests.GET("/:id", api.ManifestRouteApi.Get)
	manifests.GET("/:id/resources", api.ManifestRouteApi.GetResources)
	manifests.DELETE("/:id", api.ManifestRouteApi.Delete)
}
