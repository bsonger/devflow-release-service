package model

type ManifestRegistryConfig struct {
	Registry   string
	Namespace  string
	Repository string
	Username   string
	Password   string
	PlainHTTP  bool
}

func (c ManifestRegistryConfig) RepositoryPrefix() string {
	registry := ImageRegistryConfig{
		Registry:  c.Registry,
		Namespace: c.Namespace,
	}.Repository()
	repository := normalizeImageSegment(c.Repository)
	if repository == "" {
		repository = "manifests"
	}
	if registry == "" {
		return repository
	}
	return registry + "/" + repository
}

func (c ManifestRegistryConfig) RepositoryFor(applicationName, environmentID string) string {
	prefix := c.RepositoryPrefix()
	application := normalizeImageSegment(applicationName)
	if application == "" {
		application = "application"
	}
	environment := normalizeImageSegment(environmentID)
	if environment == "" {
		environment = "environment"
	}
	return prefix + "/" + application + "/" + environment
}
