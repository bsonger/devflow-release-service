package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"sigs.k8s.io/yaml"
)

var ErrManifestImageNotDeployable = errors.New("image has neither digest nor tag")

func resolveWorkloadImageRef(repository, tag, digest string) (string, map[string]string, error) {
	annotations := map[string]string{}
	if digest != "" {
		if tag != "" {
			annotations["devflow.io/image-tag"] = tag
			annotations["devflow.io/image-ref"] = repository + ":" + tag
		}
		return repository + "@" + digest, annotations, nil
	}
	if tag != "" {
		return repository + ":" + tag, annotations, nil
	}
	return "", nil, ErrManifestImageNotDeployable
}

func joinRenderedYAML(objects []model.ManifestRenderedObject) string {
	parts := make([]string, 0, len(objects))
	for _, item := range objects {
		if trimmed := strings.TrimSpace(item.YAML); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, "\n---\n")
}

func renderManifestObjects(namespace, applicationName, configMapName string, appConfig model.ManifestAppConfig, workload model.ManifestWorkloadConfig, services []model.ManifestService, routes []model.ManifestRoute, imageRef string, annotations map[string]string) ([]model.ManifestRenderedObject, error) {
	objects := make([]model.ManifestRenderedObject, 0, len(services)+3)

	configMap := map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":      configMapName,
			"namespace": namespace,
		},
		"data": appConfig.Data,
	}
	item, err := marshalRenderedObject("ConfigMap", configMapName, namespace, configMap)
	if err != nil {
		return nil, err
	}
	objects = append(objects, item)

	for _, service := range services {
		ports := make([]map[string]any, 0, len(service.Ports))
		for _, port := range service.Ports {
			ports = append(ports, map[string]any{
				"name":       port.Name,
				"port":       port.ServicePort,
				"targetPort": port.TargetPort,
				"protocol":   defaultProtocol(port.Protocol),
			})
		}
		serviceObj := map[string]any{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]any{
				"name":      service.Name,
				"namespace": namespace,
			},
			"spec": map[string]any{
				"selector": map[string]string{
					"app.kubernetes.io/name": applicationName,
				},
				"ports": ports,
			},
		}
		item, err := marshalRenderedObject("Service", service.Name, namespace, serviceObj)
		if err != nil {
			return nil, err
		}
		objects = append(objects, item)
	}

	httpRoutes := make([]map[string]any, 0, len(routes))
	for _, route := range routes {
		httpRoute := map[string]any{
			"name": route.Name,
			"match": []map[string]any{{
				"uri": map[string]any{"prefix": defaultPath(route.Path)},
			}},
			"route": []map[string]any{{
				"destination": map[string]any{
					"host": route.ServiceName,
					"port": map[string]any{"number": route.ServicePort},
				},
			}},
		}
		httpRoutes = append(httpRoutes, httpRoute)
	}
	hosts := uniqueHosts(routes)
	if len(hosts) > 0 && len(httpRoutes) > 0 {
		virtualServiceObj := map[string]any{
			"apiVersion": "networking.istio.io/v1beta1",
			"kind":       "VirtualService",
			"metadata": map[string]any{
				"name":      applicationName,
				"namespace": namespace,
			},
			"spec": map[string]any{
				"hosts": hosts,
				"http":  httpRoutes,
			},
		}
		item, err = marshalRenderedObject("VirtualService", applicationName, namespace, virtualServiceObj)
		if err != nil {
			return nil, err
		}
		objects = append(objects, item)
	}

	env := make([]map[string]any, 0, len(workload.Env))
	for _, entry := range workload.Env {
		env = append(env, map[string]any{"name": entry.Name, "value": entry.Value})
	}
	templateAnnotations := map[string]any{}
	for k, v := range annotations {
		templateAnnotations[k] = v
	}
	deploymentConfigName := workloadConfigResourceName(applicationName)
	deploymentObj := map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			"name":      applicationName,
			"namespace": namespace,
		},
		"spec": map[string]any{
			"replicas": workload.Replicas,
			"selector": map[string]any{
				"matchLabels": map[string]string{
					"app.kubernetes.io/name": applicationName,
				},
			},
			"template": map[string]any{
				"metadata": map[string]any{
					"labels": map[string]string{
						"app.kubernetes.io/name": applicationName,
					},
					"annotations": templateAnnotations,
				},
				"spec": map[string]any{
					"imagePullSecrets": []map[string]any{{"name": "aliyun-docker-config"}},
					"containers": []map[string]any{{
						"name":      applicationName,
						"image":     imageRef,
						"env":       env,
						"envFrom":   []map[string]any{{"configMapRef": map[string]any{"name": configMapName}}},
						"resources": workload.Resources,
						"volumeMounts": []map[string]any{{
							"name":      "config",
							"mountPath": "/etc/devflow/config/config.yaml",
							"subPath":   "config.yaml",
							"readOnly":  true,
						}},
					}},
					"volumes": []map[string]any{{
						"name": "config",
						"configMap": map[string]any{
							"name": deploymentConfigName,
						},
					}},
				},
			},
		},
	}
	if len(workload.Probes) > 0 {
		container := deploymentObj["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"].([]map[string]any)[0]
		for k, v := range workload.Probes {
			container[k] = v
		}
	}
	item, err = marshalRenderedObject("Deployment", applicationName, namespace, deploymentObj)
	if err != nil {
		return nil, err
	}
	objects = append(objects, item)
	return objects, nil
}

func marshalRenderedObject(kind, name, namespace string, object any) (model.ManifestRenderedObject, error) {
	body, err := yaml.Marshal(object)
	if err != nil {
		return model.ManifestRenderedObject{}, fmt.Errorf("marshal %s %s: %w", kind, name, err)
	}
	return model.ManifestRenderedObject{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
		YAML:      string(body),
	}, nil
}

func workloadConfigResourceName(applicationName string) string {
	trimmed := strings.TrimSpace(applicationName)
	trimmed = strings.TrimPrefix(trimmed, "devflow-")
	if trimmed == "" {
		trimmed = applicationName
	}
	return trimmed + "-config"
}

func uniqueHosts(routes []model.ManifestRoute) []string {
	seen := make(map[string]struct{}, len(routes))
	out := make([]string, 0, len(routes))
	for _, route := range routes {
		if route.Host == "" {
			continue
		}
		if _, ok := seen[route.Host]; ok {
			continue
		}
		seen[route.Host] = struct{}{}
		out = append(out, route.Host)
	}
	return out
}

func defaultPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return "/"
	}
	return path
}

func defaultProtocol(protocol string) string {
	if strings.TrimSpace(protocol) == "" {
		return "TCP"
	}
	return protocol
}
