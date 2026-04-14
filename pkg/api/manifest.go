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
	CreateManifest(context.Context, *model.CreateManifestRequest) (*model.Manifest, error)
	List(context.Context, model.ManifestListFilter) ([]model.Manifest, error)
	Get(context.Context, uuid.UUID) (*model.Manifest, error)
}

type ManifestHandler struct {
	svc manifestService
}

type ManifestResponse struct {
	Data *ManifestDoc `json:"data"`
}

type ManifestListResponse struct {
	Data       []ManifestDoc    `json:"data"`
	Pagination httpx.Pagination `json:"pagination"`
}

func NewManifestHandler() *ManifestHandler {
	return &ManifestHandler{svc: service.ManifestService}
}

// CreateManifest godoc
// @Summary Create manifest
// @Tags Manifest
// @Accept json
// @Produce json
// @Param data body api.CreateManifestRequestDoc true "Manifest create request"
// @Success 201 {object} api.ManifestResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 409 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/v1/manifests [post]
func (h *ManifestHandler) Create(c *gin.Context) {
	var req model.CreateManifestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}
	item, err := h.svc.CreateManifest(c.Request.Context(), &req)
	if err != nil {
		writeManifestError(c, err)
		return
	}
	httpx.WriteData(c, http.StatusCreated, item)
}

// ListManifests godoc
// @Summary List manifests
// @Tags Manifest
// @Produce json
// @Param application_id query string false "Application ID"
// @Param environment_id query string false "Environment ID"
// @Param image_id query string false "Image ID"
// @Param page query int false "Page"
// @Param page_size query int false "Page size"
// @Success 200 {object} api.ManifestListResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/v1/manifests [get]
func (h *ManifestHandler) List(c *gin.Context) {
	filter := model.ManifestListFilter{IncludeDeleted: httpx.IncludeDeleted(c)}
	if value := c.Query("application_id"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid application_id", nil)
			return
		}
		filter.ApplicationID = &id
	}
	if value := c.Query("environment_id"); value != "" {
		filter.EnvironmentID = &value
	}
	if value := c.Query("image_id"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid image_id", nil)
			return
		}
		filter.ImageID = &id
	}
	items, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}
	paging, err := httpx.ParsePagination(c)
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}
	total := len(items)
	items = httpx.PaginateSlice(items, paging)
	httpx.WriteList(c, http.StatusOK, items, paging, total)
}

// GetManifest godoc
// @Summary Get manifest
// @Tags Manifest
// @Produce json
// @Param id path string true "Manifest ID"
// @Success 200 {object} api.ManifestResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/v1/manifests/{id} [get]
func (h *ManifestHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid id", nil)
		return
	}
	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(c, http.StatusNotFound, "not_found", "not found", nil)
			return
		}
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
		return
	}
	httpx.WriteData(c, http.StatusOK, item)
}

func writeManifestError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		httpx.WriteError(c, http.StatusNotFound, "not_found", "not found", nil)
	case errors.Is(err, service.ErrManifestImageApplicationMismatch),
		errors.Is(err, service.ErrManifestEnvironmentBindingMissing),
		errors.Is(err, service.ErrManifestAppConfigMissing),
		errors.Is(err, service.ErrManifestWorkloadConfigMissing),
		errors.Is(err, service.ErrManifestRouteTargetInvalid),
		errors.Is(err, service.ErrManifestImageNotDeployable):
		httpx.WriteError(c, http.StatusConflict, "failed_precondition", err.Error(), nil)
	default:
		httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
	}
}
