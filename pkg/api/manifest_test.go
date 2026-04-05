package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubManifestService struct {
	createFn func(context.Context, *model.Manifest) (uuid.UUID, error)
	listFn   func(context.Context, service.ManifestListFilter) ([]model.Manifest, error)
	getFn    func(context.Context, uuid.UUID) (*model.Manifest, error)
	patchFn  func(context.Context, uuid.UUID, *model.PatchManifestRequest) error
}

func (s stubManifestService) CreateManifest(ctx context.Context, m *model.Manifest) (uuid.UUID, error) {
	return s.createFn(ctx, m)
}

func (s stubManifestService) List(ctx context.Context, filter service.ManifestListFilter) ([]model.Manifest, error) {
	return s.listFn(ctx, filter)
}

func (s stubManifestService) Get(ctx context.Context, id uuid.UUID) (*model.Manifest, error) {
	return s.getFn(ctx, id)
}

func (s stubManifestService) Patch(ctx context.Context, id uuid.UUID, patch *model.PatchManifestRequest) error {
	return s.patchFn(ctx, id, patch)
}

func TestCreateManifestReturnsEnvelope(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ManifestHandler{
		svc: stubManifestService{
			createFn: func(_ context.Context, m *model.Manifest) (uuid.UUID, error) {
				m.WithCreateDefault()
				m.Name = "manifest-v1"
				return m.GetID(), nil
			},
		},
	}

	r := gin.New()
	r.POST("/api/v1/manifests", handler.Create)

	body := bytes.NewBufferString(`{"application_id":"11111111-1111-1111-1111-111111111111","branch":"main"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/manifests", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d want %d", rec.Code, http.StatusCreated)
	}

	var payload struct {
		Data model.Manifest `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if payload.Data.ApplicationID == uuid.Nil || payload.Data.Branch != "main" {
		t.Fatalf("unexpected payload: %#v", payload.Data)
	}
}

func TestPatchManifestNotFoundReturnsErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ManifestHandler{
		svc: stubManifestService{
			patchFn: func(_ context.Context, _ uuid.UUID, _ *model.PatchManifestRequest) error {
				return sql.ErrNoRows
			},
		},
	}

	r := gin.New()
	r.PATCH("/api/v1/manifests/:id", handler.Patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/manifests/"+uuid.New().String(), bytes.NewBufferString(`{"digest":"sha256:1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d want %d", rec.Code, http.StatusNotFound)
	}
}
