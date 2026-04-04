package router

import (
	_ "github.com/bsonger/devflow-release-service/docs" // swagger docs 自动生成
	"github.com/bsonger/devflow-release-service/pkg/telemetry"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"net/http"
	"time"
)

type Module string

const (
	ModuleProject       Module = "project"
	ModuleApplication   Module = "application"
	ModuleManifest      Module = "manifest"
	ModuleJob           Module = "job"
	ModuleIntent        Module = "intent"
	ModuleConfiguration Module = "configuration"
	ModuleVerify        Module = "verify"
)

type Options struct {
	ServiceName                 string
	EnableSwagger               bool
	IncludeNestedManifestRoutes bool
	Modules                     []Module
}

// NewRouter creates the main Gin router.
func NewRouter() *gin.Engine {
	return NewRouterWithOptions(Options{
		ServiceName:                 "devflow",
		EnableSwagger:               true,
		IncludeNestedManifestRoutes: true,
		Modules: []Module{
			ModuleProject,
			ModuleApplication,
			ModuleManifest,
			ModuleJob,
			ModuleIntent,
		},
	})
}

func NewRouterWithOptions(opts Options) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New() // ⭐ 不使用 gin.Default()

	var myFilter otelgin.Filter = func(req *http.Request) bool {
		path := req.URL.Path
		return !shouldIgnore(path)
	}

	r.Use(
		otelgin.Middleware(serviceName(opts), otelgin.WithFilter(myFilter)),
		LoggerMiddleware(),
		GinZapRecovery(),
		PyroscopeMiddleware(),
		GinMetricsMiddleware(),
		GinZapLogger(),
		cors.New(cors.Config{
			AllowOrigins:     []string{"*"}, // 允许所有来源
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"*"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
	)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": serviceName(opts),
			"status":  "ok",
		})
	})

	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": serviceName(opts),
			"status":  "ready",
		})
	})

	if opts.EnableSwagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := r.Group("/api/v1")

	registerModules(api, opts)
	return r
}

func serviceName(opts Options) string {
	if opts.ServiceName == "" {
		return "devflow"
	}
	return opts.ServiceName
}

func registerModules(api *gin.RouterGroup, opts Options) {
	seen := make(map[Module]struct{}, len(opts.Modules))

	for _, module := range opts.Modules {
		if _, ok := seen[module]; ok {
			continue
		}
		seen[module] = struct{}{}

		switch module {
		case ModuleProject:
			RegisterProjectRoutes(api)
		case ModuleApplication:
			if opts.IncludeNestedManifestRoutes {
				RegisterApplicationRoutes(api)
			} else {
				RegisterApplicationCoreRoutes(api)
			}
		case ModuleManifest:
			RegisterManifestRoutes(api)
		case ModuleJob:
			RegisterJobRoutes(api)
		case ModuleIntent:
			RegisterIntentRoutes(api)
		case ModuleConfiguration:
			RegisterConfigurationRoutes(api)
		case ModuleVerify:
			RegisterVerifyRoutes(api)
		}
	}
}

func StartMetricsServer(addr string) {
	telemetry.StartMetricsServer(addr)
}
