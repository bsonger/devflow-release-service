package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubManifestService struct {
	createFn       func(context.Context, *model.CreateManifestRequest) (*model.Manifest, error)
	listFn         func(context.Context, model.ManifestListFilter) ([]model.Manifest, error)
	getFn          func(context.Context, uuid.UUID) (*model.Manifest, error)
	getResourcesFn func(context.Context, uuid.UUID) (*model.ManifestResourcesView, error)
}

func (s stubManifestService) CreateManifest(ctx context.Context, req *model.CreateManifestRequest) (*model.Manifest, error) {
	return s.createFn(ctx, req)
}

func (s stubManifestService) List(ctx context.Context, filter model.ManifestListFilter) ([]model.Manifest, error) {
	return s.listFn(ctx, filter)
}

func (s stubManifestService) Get(ctx context.Context, id uuid.UUID) (*model.Manifest, error) {
	return s.getFn(ctx, id)
}

func (s stubManifestService) GetResources(ctx context.Context, id uuid.UUID) (*model.ManifestResourcesView, error) {
	return s.getResourcesFn(ctx, id)
}

func TestCreateManifestReturnsCreated(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ManifestHandler{
		svc: stubManifestService{
			createFn: func(_ context.Context, req *model.CreateManifestRequest) (*model.Manifest, error) {
				now := mustTime("2026-04-12T11:30:00Z")
				item := &model.Manifest{ApplicationID: req.ApplicationID, EnvironmentID: req.EnvironmentID, ImageID: req.ImageID, ImageRef: "repo/demo@sha256:abc", ArtifactRepository: "repo/manifests/demo", ArtifactTag: "manifest-tag", ArtifactRef: "repo/manifests/demo:manifest-tag", ArtifactDigest: "sha256:def", ArtifactMediaType: "application/vnd.oci.image.manifest.v1+json", ArtifactPushedAt: &now, RenderedYAML: "apiVersion: v1", Status: model.ManifestReady}
				item.WithCreateDefault()
				return item, nil
			},
		},
	}
	r := gin.New()
	r.POST("/api/v1/manifests", handler.Create)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/manifests", bytes.NewBufferString(`{"application_id":"11111111-1111-1111-1111-111111111111","environment_id":"22222222-2222-2222-2222-222222222222","image_id":"33333333-3333-3333-3333-333333333333"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d want %d body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var payload struct {
		Data model.Manifest `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if payload.Data.ImageRef == "" || payload.Data.RenderedYAML == "" || payload.Data.ArtifactRef == "" || payload.Data.ArtifactDigest == "" {
		t.Fatalf("unexpected payload %+v", payload.Data)
	}
}

func TestCreateManifestAcceptsNamedEnvironmentID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ManifestHandler{
		svc: stubManifestService{
			createFn: func(_ context.Context, req *model.CreateManifestRequest) (*model.Manifest, error) {
				if req.EnvironmentID != "staging" {
					t.Fatalf("EnvironmentID = %q, want staging", req.EnvironmentID)
				}
				now := mustTime("2026-04-13T15:00:00Z")
				item := &model.Manifest{
					ApplicationID:      req.ApplicationID,
					EnvironmentID:      req.EnvironmentID,
					ImageID:            req.ImageID,
					ImageRef:           "repo/demo@sha256:abc",
					ArtifactRepository: "repo/manifests/demo/staging",
					ArtifactTag:        "demo-20260413-150000",
					ArtifactRef:        "repo/manifests/demo/staging:demo-20260413-150000",
					ArtifactDigest:     "sha256:def",
					ArtifactMediaType:  "application/vnd.oci.image.manifest.v1+json",
					ArtifactPushedAt:   &now,
					RenderedYAML:       "apiVersion: v1",
					Status:             model.ManifestReady,
				}
				item.WithCreateDefault()
				return item, nil
			},
		},
	}
	r := gin.New()
	r.POST("/api/v1/manifests", handler.Create)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/manifests", bytes.NewBufferString(`{"application_id":"11111111-1111-1111-1111-111111111111","environment_id":"staging","image_id":"33333333-3333-3333-3333-333333333333"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d want %d body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var payload struct {
		Data model.Manifest `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if payload.Data.EnvironmentID != "staging" {
		t.Fatalf("EnvironmentID = %q, want staging", payload.Data.EnvironmentID)
	}
}

func mustTime(value string) time.Time {
	got, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return got
}

func TestGetManifestNotFound(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ManifestHandler{
		svc: stubManifestService{
			getFn: func(_ context.Context, _ uuid.UUID) (*model.Manifest, error) { return nil, sql.ErrNoRows },
		},
	}
	r := gin.New()
	r.GET("/api/v1/manifests/:id", handler.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/manifests/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetManifestResourcesReturnsGroupedFrozenObjects(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	manifestID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	handler := &ManifestHandler{
		svc: stubManifestService{
			getResourcesFn: func(_ context.Context, id uuid.UUID) (*model.ManifestResourcesView, error) {
				if id != manifestID {
					t.Fatalf("id = %s want %s", id, manifestID)
				}
				return &model.ManifestResourcesView{
					ManifestID:     manifestID,
					ApplicationID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
					EnvironmentID:  "staging",
					Resources: model.ManifestGroupedResources{
						ConfigMap: &model.ManifestRenderedResource{
							Kind:      "ConfigMap",
							Name:      "demo-api-config",
							Namespace: "staging",
							YAML:      "apiVersion: v1",
							Object:    map[string]any{"kind": "ConfigMap"},
						},
						Deployment: &model.ManifestRenderedResource{
							Kind:      "Deployment",
							Name:      "demo-api",
							Namespace: "staging",
							YAML:      "apiVersion: apps/v1",
							Object:    map[string]any{"kind": "Deployment"},
						},
						Services: []model.ManifestRenderedResource{{
							Kind:      "Service",
							Name:      "demo-api",
							Namespace: "staging",
							YAML:      "apiVersion: v1",
							Object:    map[string]any{"kind": "Service"},
						}},
					},
				}, nil
			},
		},
	}
	r := gin.New()
	r.GET("/api/v1/manifests/:id/resources", handler.GetResources)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/manifests/"+manifestID.String()+"/resources", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload struct {
		Data model.ManifestResourcesView `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if payload.Data.Resources.ConfigMap == nil || payload.Data.Resources.Deployment == nil {
		t.Fatalf("unexpected grouped resources %+v", payload.Data.Resources)
	}
	if payload.Data.Resources.Rollout != nil {
		t.Fatalf("expected rollout nil, got %+v", payload.Data.Resources.Rollout)
	}
	if len(payload.Data.Resources.Services) != 1 {
		t.Fatalf("services len = %d want 1", len(payload.Data.Resources.Services))
	}
}

func TestCreateManifestClusterNotReadyReturns409(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ManifestHandler{
		svc: stubManifestService{
			createFn: func(_ context.Context, _ *model.CreateManifestRequest) (*model.Manifest, error) {
				return nil, service.ErrDeployTargetClusterNotReady
			},
		},
	}

	r := gin.New()
	r.POST("/api/v1/manifests", handler.Create)

	body := bytes.NewBufferString(`{"application_id":"11111111-1111-1111-1111-111111111111","environment_id":"env-1","image_id":"33333333-3333-3333-3333-333333333333"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/manifests", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("got %d want %d for cluster not ready", rec.Code, http.StatusConflict)
	}
	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if resp.Error.Code != "failed_precondition" {
		t.Fatalf("error code = %q, want failed_precondition", resp.Error.Code)
	}
}

func TestCreateManifestClusterReadinessMalformedReturns409(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ManifestHandler{
		svc: stubManifestService{
			createFn: func(_ context.Context, _ *model.CreateManifestRequest) (*model.Manifest, error) {
				return nil, service.ErrDeployTargetClusterReadinessMalformed
			},
		},
	}

	r := gin.New()
	r.POST("/api/v1/manifests", handler.Create)

	body := bytes.NewBufferString(`{"application_id":"11111111-1111-1111-1111-111111111111","environment_id":"env-1","image_id":"33333333-3333-3333-3333-333333333333"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/manifests", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("got %d want %d for readiness malformed", rec.Code, http.StatusConflict)
	}
	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if resp.Error.Code != "failed_precondition" {
		t.Fatalf("error code = %q, want failed_precondition", resp.Error.Code)
	}
}

func TestCreateManifestClusterNotReadyDoesNotReturnInternal500(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ManifestHandler{
		svc: stubManifestService{
			createFn: func(_ context.Context, _ *model.CreateManifestRequest) (*model.Manifest, error) {
				return nil, service.ErrDeployTargetClusterNotReady
			},
		},
	}

	r := gin.New()
	r.POST("/api/v1/manifests", handler.Create)

	body := bytes.NewBufferString(`{"application_id":"11111111-1111-1111-1111-111111111111","environment_id":"env-1","image_id":"33333333-3333-3333-3333-333333333333"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/manifests", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code == http.StatusInternalServerError {
		t.Fatalf("readiness blocker must not surface as 500 internal, got %d", rec.Code)
	}
}

func TestCreateManifestBindingMissingReturns409(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	handler := &ManifestHandler{
		svc: stubManifestService{
			createFn: func(_ context.Context, _ *model.CreateManifestRequest) (*model.Manifest, error) {
				return nil, service.ErrDeployTargetBindingMissing
			},
		},
	}

	r := gin.New()
	r.POST("/api/v1/manifests", handler.Create)

	body := bytes.NewBufferString(`{"application_id":"11111111-1111-1111-1111-111111111111","environment_id":"env-1","image_id":"33333333-3333-3333-3333-333333333333"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/manifests", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("got %d want %d for binding missing", rec.Code, http.StatusConflict)
	}
	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if resp.Error.Code != "failed_precondition" {
		t.Fatalf("error code = %q, want failed_precondition", resp.Error.Code)
	}
}
