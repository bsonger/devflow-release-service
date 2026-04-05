package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/bsonger/devflow-service-common/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ManifestRouteApi = NewManifestHandler()

type manifestService interface {
	CreateManifest(ctx context.Context, m *model.Manifest) (uuid.UUID, error)
	List(ctx context.Context, filter service.ManifestListFilter) ([]model.Manifest, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Manifest, error)
	Patch(ctx context.Context, id uuid.UUID, patch *model.PatchManifestRequest) error
}

type ManifestHandler struct {
	svc manifestService
}

func NewManifestHandler() *ManifestHandler {
	return &ManifestHandler{svc: service.ManifestService}
}

type CreateManifestRequest struct {
	ApplicationID           uuid.UUID  `json:"application_id"`
	ConfigurationRevisionID *uuid.UUID `json:"configuration_revision_id,omitempty"`
	RuntimeSpecRevisionID   *uuid.UUID `json:"runtime_spec_revision_id,omitempty"`
	Branch                  string     `json:"branch"`
}

// Create
// @Summary      创建 Manifest
// @Description  根据 Manifest 创建 Manifest，自动生成名称
// @Tags         Manifest
// @Accept       json
// @Produce      json
// @Param        data body api.CreateManifestRequest true "Manifest 数据（branch 必填）"
// @Success      201 {object} httpx.DataResponse[model.Manifest]
// @Router       /api/v1/manifests [post]
func (h *ManifestHandler) Create(c *gin.Context) {
	var req CreateManifestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}
	m := model.Manifest{
		ApplicationID:           req.ApplicationID,
		ConfigurationRevisionID: req.ConfigurationRevisionID,
		RuntimeSpecRevisionID:   req.RuntimeSpecRevisionID,
		Branch:                  req.Branch,
	}

	_, err := h.svc.CreateManifest(c.Request.Context(), &m)
	if err != nil {
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}

	httpx.WriteData(c, http.StatusCreated, m)
}

// List
// @Summary 获取应用列表
// @Tags    Manifest
// @Success 200 {object} httpx.ListResponse[model.Manifest]
// @Router  /api/v1/manifests [get]
func (h *ManifestHandler) List(c *gin.Context) {
	filter := service.ManifestListFilter{IncludeDeleted: httpx.IncludeDeleted(c)}
	if appID := c.Query("application_id"); appID != "" {
		id, err := uuid.Parse(appID)
		if err != nil {
			httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid application_id", nil)
			return
		}
		filter.ApplicationID = &id
	}
	if pipelineID := c.Query("pipeline_id"); pipelineID != "" {
		filter.PipelineID = pipelineID
	}
	if status := c.Query("status"); status != "" {
		filter.Status = status
	}
	if branch := c.Query("branch"); branch != "" {
		filter.Branch = branch
	}
	if name := c.Query("name"); name != "" {
		filter.Name = name
	}

	manifests, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}

	paging, err := httpx.ParsePagination(c)
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}

	total := len(manifests)
	manifests = httpx.PaginateSlice(manifests, paging)
	httpx.WriteList(c, http.StatusOK, manifests, paging, total)
}

// Get
// @Summary	获取应用
// @Tags		Manifest
// @Param		id	path		string	true	"Manifest ID"
// @Success	200	{object}	httpx.DataResponse[model.Manifest]
// @Router		/api/v1/manifests/{id} [get]
func (h *ManifestHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid id", nil)
		return
	}

	app, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(c, http.StatusNotFound, "not_found", "not found", nil)
			return
		}
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}

	httpx.WriteData(c, http.StatusOK, app)
}

// Patch
// @Summary		Patch Manifest
// @Description	部分更新 Manifest（仅支持 digest / commit_hash）
// @Tags		Manifest
// @Accept		json
// @Produce		json
// @Param		id		path		string			true	"Manifest ID"
// @Param		data	body		model.PatchManifestRequest	false	"Patch 数据"
// @Success		204
// @Router		/api/v1/manifests/{id} [patch]
func (h *ManifestHandler) Patch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid id", nil)
		return
	}

	var patch model.PatchManifestRequest
	if err := c.ShouldBindJSON(&patch); err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}

	err = h.svc.Patch(
		c.Request.Context(),
		id,
		&patch,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(c, http.StatusNotFound, "not_found", "manifest not found", nil)
			return
		}

		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}

	httpx.WriteNoContent(c)
}
