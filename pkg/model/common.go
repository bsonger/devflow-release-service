package model

type (
	Internet    string
	ReleaseType string
)

const (
	Internal Internet = "internal"
	External Internet = "external"

	Normal    ReleaseType = "normal"
	Canary    ReleaseType = "canary"
	BlueGreen ReleaseType = "blue-green"
)

type Service struct {
	Ports []Port `yaml:"ports" json:"ports"`
}

type Port struct {
	Name       string `bson:"name" json:"name"`
	Port       int    `bson:"port" json:"port"`
	TargetPort int    `bson:"target_port" json:"target_port"`
}

type ConfigMap struct {
	Name      string            `bson:"name" json:"name"`
	MountPath string            `bson:"mount_path" json:"mount_path"`
	FilesPath map[string]string `bson:"files_path" json:"files_path"`
}

type EnvVar struct {
	Name  string `bson:"name" json:"name"`
	Value string `bson:"value" json:"value"`
}
