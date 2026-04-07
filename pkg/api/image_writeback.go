package api

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/bsonger/devflow-service-common/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ObserverTokenHeader = "X-Devflow-Observer-Token"

var ObserverSharedToken string

type imageWritebackService interface {
	AssignPipelineID(ctx context.Context, imageID uuid.UUID, pipelineID string) error
	UpdateManifestStatusByID(ctx context.Context, imageID uuid.UUID, status model.ManifestStatus) error
	UpdateStepStatus(ctx context.Context, pipelineID, taskName string, status model.StepStatus, message string, start, end *time.Time) error
	BindTaskRun(ctx context.Context, pipelineID, taskName, taskRun string) error
	Get(ctx context.Context, id uuid.UUID) (*model.Manifest, error)
}

type ImageWritebackHandler struct {
	svc imageWritebackService
}

type ImageTektonStatusRequest struct {
	ImageID    string               `json:"image_id" binding:"required"`
	PipelineID string               `json:"pipeline_id,omitempty"`
	Status     model.ManifestStatus `json:"status" binding:"required"`
	Message    string               `json:"message,omitempty"`
}

type ImageTektonTaskRequest struct {
	ImageID    string           `json:"image_id" binding:"required"`
	PipelineID string           `json:"pipeline_id,omitempty"`
	TaskName   string           `json:"task_name" binding:"required"`
	TaskRun    string           `json:"task_run,omitempty"`
	Status     model.StepStatus `json:"status" binding:"required"`
	Message    string           `json:"message,omitempty"`
	StartTime  *time.Time       `json:"start_time,omitempty"`
	EndTime    *time.Time       `json:"end_time,omitempty"`
}

func NewImageWritebackHandler() *ImageWritebackHandler {
	return &ImageWritebackHandler{svc: service.ManifestService}
}

func RequireObserverToken(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		expected = strings.TrimSpace(expected)
		if expected == "" {
			c.Next()
			return
		}
		token := strings.TrimSpace(c.GetHeader(ObserverTokenHeader))
		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			httpx.WriteError(c, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

func resolvePipelineID(ctx *gin.Context, svc imageWritebackService, imageID uuid.UUID, pipelineID string) (string, error) {
	if strings.TrimSpace(pipelineID) != "" {
		return strings.TrimSpace(pipelineID), nil
	}
	image, err := svc.Get(ctx.Request.Context(), imageID)
	if err != nil {
		return "", err
	}
	if image.PipelineID == "" {
		return "", sql.ErrNoRows
	}
	return image.PipelineID, nil
}

func (h *ImageWritebackHandler) HandleTektonStatus(c *gin.Context) {
	var req ImageTektonStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}
	imageID, err := uuid.Parse(req.ImageID)
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid image_id", nil)
		return
	}
	if req.PipelineID != "" {
		if err := h.svc.AssignPipelineID(c.Request.Context(), imageID, req.PipelineID); err != nil {
			writeImageVerifyError(c, err)
			return
		}
	}
	if err := h.svc.UpdateManifestStatusByID(c.Request.Context(), imageID, normalizeManifestStatus(req.Status)); err != nil {
		writeImageVerifyError(c, err)
		return
	}
	httpx.WriteNoContent(c)
}

func (h *ImageWritebackHandler) HandleTektonTask(c *gin.Context) {
	var req ImageTektonTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		return
	}
	imageID, err := uuid.Parse(req.ImageID)
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_argument", "invalid image_id", nil)
		return
	}
	pipelineID, err := resolvePipelineID(c, h.svc, imageID, req.PipelineID)
	if err != nil {
		writeImageVerifyError(c, err)
		return
	}
	if req.TaskRun != "" {
		if err := h.svc.BindTaskRun(c.Request.Context(), pipelineID, req.TaskName, req.TaskRun); err != nil {
			writeImageVerifyError(c, err)
			return
		}
	}
	if err := h.svc.UpdateStepStatus(c.Request.Context(), pipelineID, req.TaskName, normalizeStepStatus(req.Status), req.Message, req.StartTime, req.EndTime); err != nil {
		writeImageVerifyError(c, err)
		return
	}
	httpx.WriteNoContent(c)
}

func normalizeManifestStatus(status model.ManifestStatus) model.ManifestStatus {
	switch strings.ToLower(string(status)) {
	case "pending":
		return model.ManifestPending
	case "running":
		return model.ManifestRunning
	case "succeeded":
		return model.ManifestSucceeded
	case "failed":
		return model.ManifestFailed
	default:
		return status
	}
}

func normalizeStepStatus(status model.StepStatus) model.StepStatus {
	switch strings.ToLower(string(status)) {
	case "pending":
		return model.StepPending
	case "running":
		return model.StepRunning
	case "succeeded":
		return model.StepSucceeded
	case "failed":
		return model.StepFailed
	default:
		return status
	}
}

func writeImageVerifyError(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		httpx.WriteError(c, http.StatusNotFound, "not_found", "image not found", nil)
		return
	}
	httpx.WriteError(c, http.StatusInternalServerError, "internal", err.Error(), nil)
}
