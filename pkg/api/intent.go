package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var IntentRouteApi = NewIntentHandler()

type IntentHandler struct{}

func NewIntentHandler() *IntentHandler {
	return &IntentHandler{}
}

// List
// @Summary 获取执行意图列表
// @Description 按 kind、status、resource 等维度查询 execution intents
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
	id, err := uuid.Parse(c.Param("id"))
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

func buildIntentFilter(c *gin.Context) (service.IntentListFilter, error) {
	filter := service.IntentListFilter{}

	if kind := strings.TrimSpace(c.Query("kind")); kind != "" {
		filter.Kind = string(model.IntentKind(kind))
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		filter.Status = string(model.IntentStatus(status))
	}
	if resourceType := strings.TrimSpace(c.Query("resource_type")); resourceType != "" {
		filter.ResourceType = resourceType
	}
	if claimedBy := strings.TrimSpace(c.Query("claimed_by")); claimedBy != "" {
		filter.ClaimedBy = claimedBy
	}

	if err := setUUIDFilter(&filter.ResourceID, "resource_id", c.Query("resource_id")); err != nil {
		return service.IntentListFilter{}, err
	}

	return filter, nil
}

func setUUIDFilter(target **uuid.UUID, field, raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid %s", field)
	}
	*target = &id
	return nil
}
