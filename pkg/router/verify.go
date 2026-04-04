package router

import (
	"github.com/bsonger/devflow-release-service/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterVerifyRoutes(rg *gin.RouterGroup) {
	verify := rg.Group("/verify")

	verify.GET("/healthz", api.VerifyRouteApi.Health)

	protected := verify.Group("")
	protected.Use(api.RequireVerifyToken())
	protected.POST("/argo/events", api.VerifyRouteApi.HandleArgoEvent)
	protected.POST("/release/steps", api.VerifyRouteApi.HandleReleaseStepEvent)
	protected.POST("/tekton/events", api.VerifyRouteApi.HandleTektonEvent)
	protected.POST("/tekton/steps", api.VerifyRouteApi.HandleTektonStepEvent)
}
