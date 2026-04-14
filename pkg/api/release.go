package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/runtimeclient"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/bsonger/devflow-service-common/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ReleaseRouteApi = NewReleaseHandler()

type releaseService interface {
	Create(ctx context.Context, release *model.Release) (uuid.UUID, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Release, error)
	List(ctx context.Context, filter service.ReleaseListFilter) ([]*model.Release, error)
}

type ReleaseHandler struct {
	svc releaseService
}

type ReleaseResponse struct {
	Data *ReleaseDoc `json:"data"`
}

type ReleaseListResponse struct {
	Data       []*ReleaseDoc    `json:"data"`
	Pagination httpx.Pagination `json:"pagination"`
}

func NewReleaseHandler() *ReleaseHandler {
	return &ReleaseHandler{svc: service.ReleaseService}
}

type CreateReleaseRequest struct {
	ManifestID uuid.UUID `json:"manifest_id"`
	Env        string    `json:"env,omitempty"`
	Type       string    `json:"type,omitempty"`
}

// Create
// @Summary 创建Release
// @Description 创建一个新的Release
// @Tags Release
// @Accept json
// @Produce json
// @Param data body api.CreateReleaseRequest true "Release Data"
// @Success 201 {object} api.ReleaseResponse
// @Router /api/v1/releases [post]
func (h *ReleaseHandler) Create(c *gin.Context) {
	var req CreateReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}
	release := &model.Release{
		ManifestID: req.ManifestID,
		Env:        req.Env,
		Type:       req.Type,
	}
	release.WithCreateDefault()
	_, err := h.svc.Create(c.Request.Context(), release)
	if err != nil {
		if errors.Is(err, service.ErrImageMissingRuntimeSpecRevision) || errors.Is(err, service.ErrRuntimeSpecBindingMismatch) || errors.Is(err, service.ErrReleaseManifestNotReady) || errors.Is(err, runtimeclient.ErrRuntimeServiceUnavailable) {
			httpx.WriteError(c, http.StatusConflict, "failed_precondition", err.Error(), nil)
			return
		}
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}

	httpx.WriteData(c, http.StatusCreated, release)
}

// Get
// @Summary 获取Release
// @Tags Release
// @Param id path string true "Release ID"
// @Success 200 {object} api.ReleaseResponse
// @Router /api/v1/releases/{id} [get]
func (h *ReleaseHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid id", nil)
		return
	}

	release, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(c, http.StatusNotFound, "not_found", "not found", nil)
			return
		}
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}

	httpx.WriteData(c, http.StatusOK, release)
}

// List
// @Summary 获取Release列表
// @Tags Release
// @Success 200 {object} api.ReleaseListResponse
// @Router /api/v1/releases [get]
func (h *ReleaseHandler) List(c *gin.Context) {
	filter := service.ReleaseListFilter{IncludeDeleted: httpx.IncludeDeleted(c)}
	if appID := c.Query("application_id"); appID != "" {
		id, err := uuid.Parse(appID)
		if err != nil {
			httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid application_id", nil)
			return
		}
		filter.ApplicationID = &id
	}
	if manifestID := c.Query("manifest_id"); manifestID != "" {
		id, err := uuid.Parse(manifestID)
		if err != nil {
			httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid manifest_id", nil)
			return
		}
		filter.ManifestID = &id
	}
	if imageID := c.Query("image_id"); imageID != "" {
		id, err := uuid.Parse(imageID)
		if err != nil {
			httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid image_id", nil)
			return
		}
		filter.ImageID = &id
	}
	if status := c.Query("status"); status != "" {
		filter.Status = status
	}
	if releaseType := c.Query("type"); releaseType != "" {
		filter.Type = releaseType
	}
	releases, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}

	paging, err := httpx.ParsePagination(c)
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}

	total := len(releases)
	releases = httpx.PaginateSlice(releases, paging)
	httpx.WriteList(c, http.StatusOK, releases, paging, total)
}
