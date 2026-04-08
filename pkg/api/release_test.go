package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubReleaseService struct {
	createFn func(context.Context, *model.Release) (uuid.UUID, error)
	getFn    func(context.Context, uuid.UUID) (*model.Release, error)
	listFn   func(context.Context, service.ReleaseListFilter) ([]*model.Release, error)
}

func (s stubReleaseService) Create(ctx context.Context, release *model.Release) (uuid.UUID, error) {
	return s.createFn(ctx, release)
}

func (s stubReleaseService) Get(ctx context.Context, id uuid.UUID) (*model.Release, error) {
	return s.getFn(ctx, id)
}

func (s stubReleaseService) List(ctx context.Context, filter service.ReleaseListFilter) ([]*model.Release, error) {
	return s.listFn(ctx, filter)
}

func TestCreateReleaseReturnsEnvelope(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ReleaseHandler{
		svc: stubReleaseService{
			createFn: func(_ context.Context, release *model.Release) (uuid.UUID, error) {
				release.WithCreateDefault()
				release.ApplicationID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
				release.Status = model.ReleasePending
				return release.GetID(), nil
			},
		},
	}

	r := gin.New()
	r.POST("/api/v1/releases", handler.Create)

	body := bytes.NewBufferString(`{"image_id":"22222222-2222-2222-2222-222222222222","env":"prod","type":"upgrade"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/releases", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d want %d", rec.Code, http.StatusCreated)
	}

	var payload struct {
		Data model.Release `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if payload.Data.ImageID == uuid.Nil || payload.Data.Env != "prod" {
		t.Fatalf("unexpected payload: %#v", payload.Data)
	}
}

func TestCreateReleaseFailedPreconditionReturnsErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ReleaseHandler{
		svc: stubReleaseService{
			createFn: func(_ context.Context, _ *model.Release) (uuid.UUID, error) {
				return uuid.Nil, service.ErrImageMissingRuntimeSpecRevision
			},
		},
	}

	r := gin.New()
	r.POST("/api/v1/releases", handler.Create)

	body := bytes.NewBufferString(`{"image_id":"22222222-2222-2222-2222-222222222222"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/releases", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("got %d want %d", rec.Code, http.StatusConflict)
	}
}
