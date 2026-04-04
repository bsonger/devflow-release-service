package config

import (
	"context"
	"fmt"
	"github.com/bsonger/devflow-common/client/argo"
	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/client/tekton"
	commonModel "github.com/bsonger/devflow-common/model"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/store"
	"github.com/bsonger/devflow-release-service/pkg/telemetry"
	"net/http"
	"strings"

	"github.com/spf13/viper"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type Config struct {
	Server    *model.ServerConfig `mapstructure:"server" json:"server" yaml:"server"`
	Mongo     *model.MongoConfig  `mapstructure:"mongo"  json:"mongo"  yaml:"mongo"`
	Log       *model.LogConfig    `mapstructure:"log"    json:"log"    yaml:"log"`
	Otel      *model.OtelConfig   `mapstructure:"otel"   json:"otel"   yaml:"otel"`
	Repo      *model.Repo         `mapstructure:"repo"   json:"repo"   yaml:"repo"`
	Consul    *model.Consul       `mapstructure:"consul" json:"consul" yaml:"consul"`
	Pyroscope string              `mapstructure:"pyroscope" json:"pyroscope" yaml:"pyroscope"`
}

func Load() (*Config, error) {
	v := viper.New()
	//v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config/")
	v.AddConfigPath("/etc/devflow/config/")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config *Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}
	var err error
	model.KubeConfig, err = LoadKubeConfig()
	if err != nil {
		return nil, err
	}
	//err = consul.InitConsulClient(config.Consul)
	//if err != nil {
	//	return nil, err
	//}
	//consul.LoadConsulConfigAndMerge(config.Consul)

	return config, nil
}

func InitConfig(ctx context.Context, config *Config) error {
	_, err := InitRuntime(ctx, config, "")
	return err
}

func InitRuntime(ctx context.Context, config *Config, serviceName string) (func(context.Context) error, error) {
	shutdown, err := telemetry.Init(ctx, config.Log, config.Otel, config.Pyroscope, serviceName)
	if err != nil {
		return nil, err
	}

	client, err := mongo.InitMongo(ctx, toCommonMongoConfig(config.Mongo), logging.Logger)
	if err != nil {
		return shutdown, err
	}
	store.InitMongo(client, config.Mongo.DBName)
	kubeconfig, err := LoadKubeConfig()
	if err != nil {
		return shutdown, err
	}
	err = tekton.InitTektonClient(ctx, kubeconfig, logging.Logger)
	if err != nil {
		return shutdown, err
	}
	err = argo.InitArgoCdClient(kubeconfig)
	if err != nil {
		return shutdown, err
	}
	model.InitConfigRepo(config.Repo)
	return shutdown, nil
}

func toCommonMongoConfig(cfg *model.MongoConfig) *commonModel.MongoConfig {
	if cfg == nil {
		return nil
	}
	return &commonModel.MongoConfig{
		URI:    cfg.URI,
		DBName: cfg.DBName,
	}
}

func LoadKubeConfig() (*rest.Config, error) {
	// 1️⃣ 尝试本地 kubeconfig
	if cfg, err := loadLocalKubeConfig(); err == nil {
		cfg.WrapTransport = wrapK8sTransport()
		return cfg, nil
	}

	// 2️⃣ 回退到 InCluster
	if cfg, err := rest.InClusterConfig(); err == nil {
		cfg.WrapTransport = wrapK8sTransport()
		return cfg, nil
	}

	return nil, fmt.Errorf("failed to load kubeconfig (local & in-cluster)")
}

// loadLocalKubeConfig 从 $HOME/.kube/config 加载
func loadLocalKubeConfig() (*rest.Config, error) {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows fallback
	}

	kubeconfig := filepath.Join(home, ".kube", "config")

	// 如果文件不存在，直接返回 error
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return nil, err
	}

	// 使用 kubeconfig 构建 config
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func wrapK8sTransport() func(http.RoundTripper) http.RoundTripper {
	return func(rt http.RoundTripper) http.RoundTripper {
		return otelhttp.NewTransport(
			rt,
			otelhttp.WithTracerProvider(otel.GetTracerProvider()),
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				// 更清晰的 span 名称
				return fmt.Sprintf("k8s.api %s %s", r.Method, r.URL.Path)
			}),
			otelhttp.WithFilter(func(r *http.Request) bool {
				if r.Method == http.MethodPost &&
					strings.HasSuffix(r.URL.Path, "/pipelineruns") {
					return false
				}
				return true
			}),
		)
	}
}
