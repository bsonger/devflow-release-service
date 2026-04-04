package api

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var VerifyRouteApi = NewVerifyHandler()

type VerifyHandler struct{}

const VerifyTokenHeader = "X-Devflow-Verify-Token"

type VerifyBuildStatusRequest struct {
	IntentID    string               `json:"intent_id,omitempty"`
	ManifestID  string               `json:"manifest_id" binding:"required"`
	PipelineID  string               `json:"pipeline_id,omitempty"`
	Status      model.ManifestStatus `json:"status" binding:"required"`
	Message     string               `json:"message,omitempty"`
	ExternalRef string               `json:"external_ref,omitempty"`
}

type VerifyReleaseStatusRequest struct {
	IntentID    string          `json:"intent_id,omitempty"`
	JobID       string          `json:"job_id" binding:"required"`
	Status      model.JobStatus `json:"status" binding:"required"`
	Message     string          `json:"message,omitempty"`
	ExternalRef string          `json:"external_ref,omitempty"`
}

type VerifyReleaseStepRequest struct {
	JobID     string           `json:"job_id" binding:"required"`
	StepName  string           `json:"step_name" binding:"required"`
	Status    model.StepStatus `json:"status" binding:"required"`
	Progress  int32            `json:"progress,omitempty"`
	Message   string           `json:"message,omitempty"`
	StartTime *time.Time       `json:"start_time,omitempty"`
	EndTime   *time.Time       `json:"end_time,omitempty"`
}

type VerifyBuildStepRequest struct {
	ManifestID string           `json:"manifest_id" binding:"required"`
	PipelineID string           `json:"pipeline_id,omitempty"`
	TaskName   string           `json:"task_name" binding:"required"`
	TaskRun    string           `json:"task_run,omitempty"`
	Status     model.StepStatus `json:"status" binding:"required"`
	Message    string           `json:"message,omitempty"`
	StartTime  *time.Time       `json:"start_time,omitempty"`
	EndTime    *time.Time       `json:"end_time,omitempty"`
}

func NewVerifyHandler() *VerifyHandler {
	return &VerifyHandler{}
}

func RequireVerifyToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := strings.TrimSpace(os.Getenv("VERIFY_SERVICE_SHARED_TOKEN"))
		if expected == "" {
			c.Next()
			return
		}

		token := strings.TrimSpace(c.GetHeader(VerifyTokenHeader))
		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Next()
	}
}

// Health
// @Summary Verify Service 健康检查
// @Tags Verify
// @Success 200 {object} map[string]string
// @Router /api/v1/verify/healthz [get]
func (h *VerifyHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "verify-service",
		"status":  "ok",
	})
}

// HandleArgoEvent
// @Summary 回写发布状态
// @Description 由 Argo 或外部发布观察器回写 Job 级状态
// @Tags Verify
// @Accept json
// @Produce json
// @Param data body VerifyReleaseStatusRequest true "Release Status Data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/verify/argo/events [post]
func (h *VerifyHandler) HandleArgoEvent(c *gin.Context) {
	var req VerifyReleaseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID, err := primitive.ObjectIDFromHex(req.JobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	if err := service.JobService.UpdateStatus(c.Request.Context(), jobID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.IntentID != "" {
		if intentID, err := primitive.ObjectIDFromHex(req.IntentID); err == nil {
			_ = service.IntentService.UpdateStatus(c.Request.Context(), intentID, mapJobStatusToIntentStatus(req.Status), req.ExternalRef, req.Message)
		}
	} else {
		_ = service.IntentService.UpdateStatusByResource(c.Request.Context(), model.IntentKindRelease, jobID, mapJobStatusToIntentStatus(req.Status), req.ExternalRef, req.Message)
	}

	c.JSON(http.StatusOK, gin.H{"message": "release status updated"})
}

// HandleTektonEvent
// @Summary 回写构建状态
// @Description 由 Tekton 或外部构建观察器回写 Manifest 级状态
// @Tags Verify
// @Accept json
// @Produce json
// @Param data body VerifyBuildStatusRequest true "Build Status Data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/verify/tekton/events [post]
func (h *VerifyHandler) HandleTektonEvent(c *gin.Context) {
	var req VerifyBuildStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	manifestID, err := primitive.ObjectIDFromHex(req.ManifestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manifest_id"})
		return
	}

	if req.PipelineID != "" {
		if err := service.ManifestService.AssignPipelineID(c.Request.Context(), manifestID, req.PipelineID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := service.ManifestService.UpdateManifestStatusByID(c.Request.Context(), manifestID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.IntentID != "" {
		if intentID, err := primitive.ObjectIDFromHex(req.IntentID); err == nil {
			_ = service.IntentService.UpdateStatus(c.Request.Context(), intentID, mapManifestStatusToIntentStatus(req.Status), req.ExternalRef, req.Message)
		}
	} else {
		_ = service.IntentService.UpdateStatusByResource(c.Request.Context(), model.IntentKindBuild, manifestID, mapManifestStatusToIntentStatus(req.Status), req.ExternalRef, req.Message)
	}

	c.JSON(http.StatusOK, gin.H{"message": "build status updated"})
}

// HandleTektonStepEvent
// @Summary 回写构建步骤
// @Description 由 Tekton TaskRun 观察器回写 Manifest steps
// @Tags Verify
// @Accept json
// @Produce json
// @Param data body VerifyBuildStepRequest true "Build Step Data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/verify/tekton/steps [post]
func (h *VerifyHandler) HandleTektonStepEvent(c *gin.Context) {
	var req VerifyBuildStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	manifestID, err := primitive.ObjectIDFromHex(req.ManifestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manifest_id"})
		return
	}

	if req.PipelineID == "" {
		manifest, err := service.ManifestService.Get(c.Request.Context(), manifestID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		req.PipelineID = manifest.PipelineID
	}

	if req.PipelineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline_id is required until manifest is bound"})
		return
	}

	if req.TaskRun != "" {
		if err := service.ManifestService.BindTaskRun(c.Request.Context(), req.PipelineID, req.TaskName, req.TaskRun); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := service.ManifestService.UpdateStepStatus(c.Request.Context(), req.PipelineID, req.TaskName, req.Status, req.Message, req.StartTime, req.EndTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "build step updated"})
}

// HandleReleaseStepEvent
// @Summary 回写发布步骤
// @Description 由 Argo Application / Deployment / Rollout 观察器回写 Job steps
// @Tags Verify
// @Accept json
// @Produce json
// @Param data body VerifyReleaseStepRequest true "Release Step Data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/verify/release/steps [post]
func (h *VerifyHandler) HandleReleaseStepEvent(c *gin.Context) {
	var req VerifyReleaseStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID, err := primitive.ObjectIDFromHex(req.JobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	if err := service.JobService.UpdateStep(c.Request.Context(), jobID, req.StepName, req.Status, req.Progress, req.Message, req.StartTime, req.EndTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "release step updated"})
}

func mapManifestStatusToIntentStatus(status model.ManifestStatus) model.IntentStatus {
	switch status {
	case model.ManifestSucceeded:
		return model.IntentSucceeded
	case model.ManifestFailed:
		return model.IntentFailed
	default:
		return model.IntentRunning
	}
}

func mapJobStatusToIntentStatus(status model.JobStatus) model.IntentStatus {
	switch status {
	case model.JobSucceeded, model.JobRolledBack:
		return model.IntentSucceeded
	case model.JobFailed, model.JobSyncFailed:
		return model.IntentFailed
	default:
		return model.IntentRunning
	}
}
