package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var IntentRouteApi = NewIntentHandler()

type IntentHandler struct{}

func NewIntentHandler() *IntentHandler {
	return &IntentHandler{}
}

// List
// @Summary 获取执行意图列表
// @Description 按 kind、status、resource、application 等维度查询 execution intents
// @Tags Intent
// @Success 200 {array} model.Intent
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/intents [get]
func (h *IntentHandler) List(c *gin.Context) {
	filter, err := buildIntentFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	intents, err := service.IntentService.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	paging, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	total := len(intents)
	intents = paginateSlice(intents, paging)
	setPaginationHeaders(c, total, paging)

	c.JSON(http.StatusOK, intents)
}

// Get
// @Summary 获取执行意图
// @Tags Intent
// @Param id path string true "Intent ID"
// @Success 200 {object} model.Intent
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/intents/{id} [get]
func (h *IntentHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	intent, err := service.IntentService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, intent)
}

func buildIntentFilter(c *gin.Context) (primitive.M, error) {
	filter := primitive.M{}

	if kind := strings.TrimSpace(c.Query("kind")); kind != "" {
		filter["kind"] = model.IntentKind(kind)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		filter["status"] = model.IntentStatus(status)
	}
	if resourceType := strings.TrimSpace(c.Query("resource_type")); resourceType != "" {
		filter["resource_type"] = resourceType
	}
	if applicationName := strings.TrimSpace(c.Query("application_name")); applicationName != "" {
		filter["application_name"] = applicationName
	}
	if manifestName := strings.TrimSpace(c.Query("manifest_name")); manifestName != "" {
		filter["manifest_name"] = manifestName
	}
	if jobType := strings.TrimSpace(c.Query("job_type")); jobType != "" {
		filter["job_type"] = jobType
	}
	if env := strings.TrimSpace(c.Query("env")); env != "" {
		filter["env"] = env
	}
	if branch := strings.TrimSpace(c.Query("branch")); branch != "" {
		filter["branch"] = branch
	}
	if claimedBy := strings.TrimSpace(c.Query("claimed_by")); claimedBy != "" {
		filter["claimed_by"] = claimedBy
	}
	if externalRef := strings.TrimSpace(c.Query("external_ref")); externalRef != "" {
		filter["external_ref"] = externalRef
	}

	if err := setObjectIDFilter(filter, "resource_id", c.Query("resource_id")); err != nil {
		return nil, err
	}
	if err := setObjectIDFilter(filter, "application_id", c.Query("application_id")); err != nil {
		return nil, err
	}
	if err := setObjectIDFilter(filter, "manifest_id", c.Query("manifest_id")); err != nil {
		return nil, err
	}
	if err := setObjectIDFilter(filter, "job_id", c.Query("job_id")); err != nil {
		return nil, err
	}

	return filter, nil
}

func setObjectIDFilter(filter primitive.M, field, raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}

	id, err := primitive.ObjectIDFromHex(value)
	if err != nil {
		return fmt.Errorf("invalid %s", field)
	}

	filter[field] = id
	return nil
}
