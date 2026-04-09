package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/bsonger/devflow-service-common/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var IntentRouteApi = NewIntentHandler()

type intentService interface {
	Get(ctx context.Context, id uuid.UUID) (*model.Intent, error)
	List(ctx context.Context, filter service.IntentListFilter) ([]*model.Intent, error)
}

type IntentHandler struct {
	svc intentService
}

type IntentResponse struct {
	Data *model.Intent `json:"data"`
}

type IntentListResponse struct {
	Data       []*model.Intent  `json:"data"`
	Pagination httpx.Pagination `json:"pagination"`
}

func NewIntentHandler() *IntentHandler {
	return &IntentHandler{svc: service.IntentService}
}

// List
// @Summary 获取执行意图列表
// @Description 按 kind、status、resource 等维度查询 execution intents
// @Tags Intent
// @Success 200 {object} api.IntentListResponse
// @Router /api/v1/intents [get]
func (h *IntentHandler) List(c *gin.Context) {
	filter, err := buildIntentFilter(c)
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}

	intents, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}

	paging, err := httpx.ParsePagination(c)
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}

	total := len(intents)
	intents = httpx.PaginateSlice(intents, paging)
	httpx.WriteList(c, http.StatusOK, intents, paging, total)
}

// Get
// @Summary 获取执行意图
// @Tags Intent
// @Param id path string true "Intent ID"
// @Success 200 {object} api.IntentResponse
// @Router /api/v1/intents/{id} [get]
func (h *IntentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid id", nil)
		return
	}

	intent, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(c, http.StatusNotFound, "not_found", "not found", nil)
			return
		}
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}

	httpx.WriteData(c, http.StatusOK, intent)
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
