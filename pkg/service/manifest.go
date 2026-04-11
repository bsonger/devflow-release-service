package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bsonger/devflow-release-service/pkg/downstream"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/store"
	"github.com/google/uuid"
)

var (
	ErrManifestImageApplicationMismatch  = errors.New("image does not belong to application")
	ErrManifestEnvironmentBindingMissing = errors.New("environment is not bound to application")
	ErrManifestAppConfigMissing          = errors.New("effective app config is missing")
	ErrManifestWorkloadConfigMissing     = errors.New("effective workload config is missing")
	ErrManifestRouteTargetInvalid        = errors.New("route points to missing service or port")
)

var ManifestService = NewManifestService()

type manifestImageReader interface {
	Get(context.Context, uuid.UUID) (*model.Image, error)
}

type manifestOrchestratorReader interface {
	GetApplicationEnvironment(context.Context, string, string) (*downstream.ApplicationEnvironment, error)
}

type manifestNetworkReader interface {
	ListServices(context.Context, string) ([]downstream.ManifestService, error)
	ListRoutes(context.Context, string) ([]downstream.ManifestRoute, error)
}

type manifestConfigReader interface {
	FindAppConfig(context.Context, string, string) (*downstream.AppConfig, error)
	FindWorkloadConfig(context.Context, string, string) (*downstream.WorkloadConfig, error)
}

type manifestService struct {
	images       manifestImageReader
	orchestrator manifestOrchestratorReader
	networks     manifestNetworkReader
	configs      manifestConfigReader
	apps         *applicationService
}

func NewManifestService() *manifestService {
	return &manifestService{
		images:       ImageService,
		orchestrator: downstream.NewOrchestratorManifestClient(strings.TrimSpace(os.Getenv("PLATFORM_ORCHESTRATOR_BASE_URL"))),
		networks:     downstream.NewNetworkManifestClient(strings.TrimSpace(os.Getenv("NETWORK_SERVICE_BASE_URL"))),
		configs:      downstream.NewConfigManifestClient(strings.TrimSpace(os.Getenv("CONFIG_SERVICE_BASE_URL"))),
		apps:         ApplicationService,
	}
}

func (s *manifestService) CreateManifest(ctx context.Context, req *model.CreateManifestRequest) (*model.Manifest, error) {
	image, err := s.images.Get(ctx, req.ImageID)
	if err != nil {
		return nil, err
	}
	if image.ApplicationID != req.ApplicationID {
		return nil, ErrManifestImageApplicationMismatch
	}
	if _, err := s.orchestrator.GetApplicationEnvironment(ctx, req.ApplicationID.String(), req.EnvironmentID.String()); err != nil {
		return nil, ErrManifestEnvironmentBindingMissing
	}
	appConfig, err := s.configs.FindAppConfig(ctx, req.ApplicationID.String(), req.EnvironmentID.String())
	if err != nil {
		return nil, err
	}
	if appConfig == nil || (len(appConfig.Files) == 0 && len(appConfig.RenderedConfigMap) == 0) {
		return nil, ErrManifestAppConfigMissing
	}
	workloadConfig, err := s.configs.FindWorkloadConfig(ctx, req.ApplicationID.String(), req.EnvironmentID.String())
	if err != nil {
		return nil, err
	}
	if workloadConfig == nil {
		return nil, ErrManifestWorkloadConfigMissing
	}
	services, err := s.networks.ListServices(ctx, req.ApplicationID.String())
	if err != nil {
		return nil, err
	}
	routes, err := s.networks.ListRoutes(ctx, req.ApplicationID.String())
	if err != nil {
		return nil, err
	}
	application, err := s.apps.Get(ctx, req.ApplicationID)
	if err != nil {
		return nil, err
	}

	manifest, err := buildManifest(req, image, application.Name, appConfig, workloadConfig, services, routes, req.EnvironmentID.String())
	if err != nil {
		return nil, err
	}
	manifest.WithCreateDefault()
	if err := s.insert(ctx, manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (s *manifestService) List(ctx context.Context, filter model.ManifestListFilter) ([]model.Manifest, error) {
	query := `
		select id, application_id, environment_id, image_id, image_ref,
			services_snapshot, routes_snapshot, app_config_snapshot, workload_config_snapshot,
			rendered_objects, rendered_yaml, status, created_at, updated_at, deleted_at
		from manifests
	`
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	if !filter.IncludeDeleted {
		clauses = append(clauses, "deleted_at is null")
	}
	if filter.ApplicationID != nil {
		args = append(args, *filter.ApplicationID)
		clauses = append(clauses, placeholderClause("application_id", len(args)))
	}
	if filter.EnvironmentID != nil {
		args = append(args, *filter.EnvironmentID)
		clauses = append(clauses, placeholderClause("environment_id", len(args)))
	}
	if filter.ImageID != nil {
		args = append(args, *filter.ImageID)
		clauses = append(clauses, placeholderClause("image_id", len(args)))
	}
	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}
	query += " order by created_at desc"
	rows, err := store.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Manifest, 0)
	for rows.Next() {
		item, err := scanManifest(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (s *manifestService) Get(ctx context.Context, id uuid.UUID) (*model.Manifest, error) {
	return scanManifest(store.DB().QueryRowContext(ctx, `
		select id, application_id, environment_id, image_id, image_ref,
			services_snapshot, routes_snapshot, app_config_snapshot, workload_config_snapshot,
			rendered_objects, rendered_yaml, status, created_at, updated_at, deleted_at
		from manifests
		where id = $1 and deleted_at is null
	`, id))
}

func (s *manifestService) insert(ctx context.Context, m *model.Manifest) error {
	servicesJSON, err := marshalJSON(m.ServicesSnapshot, "[]")
	if err != nil {
		return err
	}
	routesJSON, err := marshalJSON(m.RoutesSnapshot, "[]")
	if err != nil {
		return err
	}
	appConfigJSON, err := marshalJSON(m.AppConfigSnapshot, "{}")
	if err != nil {
		return err
	}
	workloadJSON, err := marshalJSON(m.WorkloadConfigSnapshot, "{}")
	if err != nil {
		return err
	}
	renderedJSON, err := marshalJSON(m.RenderedObjects, "[]")
	if err != nil {
		return err
	}
	_, err = store.DB().ExecContext(ctx, `
		insert into manifests (
			id, application_id, environment_id, image_id, image_ref,
			services_snapshot, routes_snapshot, app_config_snapshot, workload_config_snapshot,
			rendered_objects, rendered_yaml, status, created_at, updated_at, deleted_at
		) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`, m.ID, m.ApplicationID, m.EnvironmentID, m.ImageID, m.ImageRef,
		servicesJSON, routesJSON, appConfigJSON, workloadJSON, renderedJSON, m.RenderedYAML,
		m.Status, m.CreatedAt, m.UpdatedAt, m.DeletedAt)
	return err
}

func buildManifest(req *model.CreateManifestRequest, image *model.Image, applicationName string, appConfig *downstream.AppConfig, workload *downstream.WorkloadConfig, services []downstream.ManifestService, routes []downstream.ManifestRoute, namespace string) (*model.Manifest, error) {
	servicesSnapshot := make([]model.ManifestService, 0, len(services))
	servicePorts := make(map[string]map[int]struct{}, len(services))
	for _, item := range services {
		ports := make([]model.ManifestServicePort, 0, len(item.Ports))
		knownPorts := make(map[int]struct{}, len(item.Ports))
		for _, port := range item.Ports {
			ports = append(ports, model.ManifestServicePort{Name: port.Name, ServicePort: port.ServicePort, TargetPort: port.TargetPort, Protocol: port.Protocol})
			knownPorts[port.ServicePort] = struct{}{}
		}
		servicesSnapshot = append(servicesSnapshot, model.ManifestService{ID: item.ID, Name: item.Name, Ports: ports})
		servicePorts[item.Name] = knownPorts
	}
	routesSnapshot := make([]model.ManifestRoute, 0, len(routes))
	for _, item := range routes {
		if _, ok := servicePorts[item.ServiceName]; !ok {
			return nil, fmt.Errorf("%w: service %s", ErrManifestRouteTargetInvalid, item.ServiceName)
		}
		if _, ok := servicePorts[item.ServiceName][item.ServicePort]; !ok {
			return nil, fmt.Errorf("%w: service %s port %d", ErrManifestRouteTargetInvalid, item.ServiceName, item.ServicePort)
		}
		routesSnapshot = append(routesSnapshot, model.ManifestRoute{
			ID: item.ID, Name: item.Name, Host: item.Host, Path: item.Path, ServiceName: item.ServiceName, ServicePort: item.ServicePort,
		})
	}
	configData := appConfig.RenderedConfigMap
	if len(configData) == 0 && len(appConfig.Files) > 0 {
		configData = make(map[string]string, len(appConfig.Files))
	}
	files := make([]model.ManifestFile, 0, len(appConfig.Files))
	for _, file := range appConfig.Files {
		files = append(files, model.ManifestFile{Name: file.Name, Content: file.Content})
		if len(configData) > 0 {
			configData[file.Name] = file.Content
		}
	}
	imageRepository := strings.TrimRight(image.RepoAddress, "/") + "/" + image.Name
	imageRef, annotations, err := resolveWorkloadImageRef(imageRepository, image.Tag, image.Digest)
	if err != nil {
		return nil, err
	}

	appConfigSnapshot := model.ManifestAppConfig{
		ID:           appConfig.ID,
		Name:         appConfig.Name,
		Files:        files,
		Data:         configData,
		SourcePath:   appConfig.SourcePath,
		SourceCommit: appConfig.SourceCommit,
	}
	workloadSnapshot := model.ManifestWorkloadConfig{
		ID:           workload.ID,
		Name:         workload.Name,
		Replicas:     workload.Replicas,
		Resources:    workload.Resources,
		Probes:       workload.Probes,
		Env:          toModelEnvVars(workload.Env),
		WorkloadType: workload.WorkloadType,
		Strategy:     workload.Strategy,
	}

	configMapName := applicationName + "-config"
	renderedObjects, err := renderManifestObjects(namespace, applicationName, configMapName, appConfigSnapshot, workloadSnapshot, servicesSnapshot, routesSnapshot, imageRef, annotations)
	if err != nil {
		return nil, err
	}
	return &model.Manifest{
		ApplicationID:          req.ApplicationID,
		EnvironmentID:          req.EnvironmentID,
		ImageID:                req.ImageID,
		ImageRef:               imageRef,
		ServicesSnapshot:       servicesSnapshot,
		RoutesSnapshot:         routesSnapshot,
		AppConfigSnapshot:      appConfigSnapshot,
		WorkloadConfigSnapshot: workloadSnapshot,
		RenderedObjects:        renderedObjects,
		RenderedYAML:           joinRenderedYAML(renderedObjects),
		Status:                 model.ManifestReady,
	}, nil
}

func toModelEnvVars(items []downstream.EnvVar) []model.EnvVar {
	out := make([]model.EnvVar, 0, len(items))
	for _, item := range items {
		out = append(out, model.EnvVar{Name: item.Name, Value: item.Value})
	}
	return out
}
