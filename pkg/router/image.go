package router

import (
	"github.com/bsonger/devflow-release-service/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterImageRoutes(rg *gin.RouterGroup) {
	images := rg.Group("/images")
	images.GET("", api.ImageRouteApi.List)
	images.POST("", api.ImageRouteApi.Create)
	images.GET("/:id", api.ImageRouteApi.Get)
	images.PATCH("/:id", api.ImageRouteApi.Patch)

	writeback := rg.Group("/images/tekton")
	writeback.Use(api.RequireObserverToken(api.ObserverSharedToken))
	handler := api.NewImageWritebackHandler()
	writeback.POST("/status", handler.HandleTektonStatus)
	writeback.POST("/tasks", handler.HandleTektonTask)
}
