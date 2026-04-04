package api

import (
	"net/http"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ReleaseRouteApi = NewReleaseHandler()

type ReleaseHandler struct{}

func NewReleaseHandler() *ReleaseHandler {
	return &ReleaseHandler{}
}

// Create
// @Summary 创建Release
// @Description 创建一个新的Release
// @Tags Release
// @Accept json
// @Produce json
// @Param data body model.Release true "Release Data"
// @Success 200 {object} CreateResponse
// @Router /api/v1/releases [post]
func (h *ReleaseHandler) Create(c *gin.Context) {
	var release *model.Release
	if err := c.ShouldBindJSON(&release); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	release.WithCreateDefault()
	id, err := service.ReleaseService.Create(c.Request.Context(), release)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newCreateResponse(id, release.ExecutionIntentID))
}

// Get
// @Summary 获取Release
// @Tags Release
// @Param id path string true "Release ID"
// @Success 200 {object} model.Release
// @Router /api/v1/releases/{id} [get]
func (h *ReleaseHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	release, err := service.ReleaseService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, release)
}

// List
// @Summary 获取Release列表
// @Tags Release
// @Success 200 {array} model.Release
// @Router /api/v1/releases [get]
func (h *ReleaseHandler) List(c *gin.Context) {
	filter := primitive.M{}
	if !includeDeleted(c) {
		filter["deleted_at"] = primitive.M{"$exists": false}
	}
	if appID := c.Query("application_id"); appID != "" {
		id, err := primitive.ObjectIDFromHex(appID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application_id"})
			return
		}
		filter["application_id"] = id
	}
	if manifestID := c.Query("manifest_id"); manifestID != "" {
		id, err := primitive.ObjectIDFromHex(manifestID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manifest_id"})
			return
		}
		filter["manifest_id"] = id
	}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	if releaseType := c.Query("type"); releaseType != "" {
		filter["type"] = releaseType
	}
	if projectName := c.Query("project_name"); projectName != "" {
		filter["project_name"] = projectName
	}
	if appName := c.Query("application_name"); appName != "" {
		filter["application_name"] = appName
	}

	releases, err := service.ReleaseService.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	paging, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	total := len(releases)
	releases = paginateSlice(releases, paging)
	setPaginationHeaders(c, total, paging)

	c.JSON(http.StatusOK, releases)
}
