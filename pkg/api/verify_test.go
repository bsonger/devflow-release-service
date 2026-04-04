package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireVerifyTokenAllowsWhenUnset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("VERIFY_SERVICE_SHARED_TOKEN", "")

	router := gin.New()
	router.Use(RequireVerifyToken())
	router.POST("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestRequireVerifyTokenRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("VERIFY_SERVICE_SHARED_TOKEN", "secret-token")

	router := gin.New()
	router.Use(RequireVerifyToken())
	router.POST("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set(VerifyTokenHeader, "wrong-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireVerifyTokenAcceptsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("VERIFY_SERVICE_SHARED_TOKEN", "secret-token")

	router := gin.New()
	router.Use(RequireVerifyToken())
	router.POST("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set(VerifyTokenHeader, "secret-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestRequireVerifyTokenUsesEnvironmentAtRequestTime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequireVerifyToken())
	router.POST("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	if err := os.Setenv("VERIFY_SERVICE_SHARED_TOKEN", "dynamic-secret"); err != nil {
		t.Fatalf("setenv failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv("VERIFY_SERVICE_SHARED_TOKEN")
	})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set(VerifyTokenHeader, "dynamic-secret")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}
}
