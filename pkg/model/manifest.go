package model

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tknv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ManifestStatus string

const (
	ManifestPending   ManifestStatus = "Pending"
	ManifestRunning   ManifestStatus = "Running"
	ManifestSucceeded ManifestStatus = "Succeeded"
	ManifestFailed    ManifestStatus = "Failed"
)

type StepStatus string

const (
	StepPending   StepStatus = "Pending"
	StepRunning   StepStatus = "Running"
	StepSucceeded StepStatus = "Succeeded"
	StepFailed    StepStatus = "Failed"
)

const (
	TraceIDAnnotation = "otel.devflow.io/trace-id"
	SpanAnnotation    = "otel.devflow.io/parent-span-id"
)

type Manifest struct {
	BaseModel         `bson:",inline"`
	ExecutionIntentID *primitive.ObjectID `json:"execution_intent_id,omitempty" bson:"execution_intent_id,omitempty"`
	ApplicationId     primitive.ObjectID  `json:"application_id" bson:"application_id"`
	Name              string              `json:"name" bson:"name"`
	ApplicationName   string              `json:"application_name" bson:"application_name"`
	Branch            string              `json:"branch" bson:"branch"`
	GitRepo           string              `json:"git_repo" bson:"git_repo"`
	CommitHash        string              `json:"commit_hash,omitempty" bson:"commit_hash,omitempty"`
	Replica           *int32              `bson:"replica" json:"replica"`
	Digest            string              `json:"digest,omitempty" bson:"digest,omitempty"`
	Type              ReleaseType         `bson:"type" json:"type"`
	ConfigMaps        []*ConfigMap        `bson:"config_maps,omitempty" json:"config_maps,omitempty"`
	Service           Service             `bson:"service" json:"service"`
	Internet          Internet            `bson:"internet" json:"internet"`
	Envs              map[string][]EnvVar `bson:"envs,omitempty" json:"envs,omitempty"`
	PipelineID        string              `json:"pipeline_id" bson:"pipeline_id"`
	Steps             []ManifestStep      `json:"steps" bson:"steps"`
	Status            ManifestStatus      `json:"status" bson:"status"`
}

type ManifestStep struct {
	TaskName  string     `bson:"task_name" json:"task_name"`
	TaskRun   string     `bson:"task_run,omitempty" json:"task_run,omitempty"`
	Status    StepStatus `bson:"status" json:"status"`
	StartTime *time.Time `bson:"start_time,omitempty" json:"start_time,omitempty"`
	EndTime   *time.Time `bson:"end_time,omitempty" json:"end_time,omitempty"`
	Message   string     `bson:"message,omitempty" json:"message,omitempty"`
}

type PatchManifestRequest struct {
	CommitHash string `json:"commit_hash,omitempty"`
	Digest     string `json:"digest,omitempty"`
}

func (r *PatchManifestRequest) IsEmpty() bool {
	return r.CommitHash == "" && r.Digest == ""
}

func GenerateManifestVersion(name string) string {
	t := time.Now().Format("20060102150405")
	r := rand.Intn(100)
	return fmt.Sprintf("%s%s%s", name, t, strconv.Itoa(r))
}

func (m *Manifest) CollectionName() string { return "manifests" }

func (m *Manifest) GetStep(taskName string) *ManifestStep {
	for i := range m.Steps {
		if m.Steps[i].TaskName == taskName {
			return &m.Steps[i]
		}
	}
	return nil
}

func (m *Manifest) GeneratePipelineRun(pipelineName string, pvc string) *tknv1.PipelineRun {
	return &tknv1.PipelineRun{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PipelineRun",
			APIVersion: "tekton.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pipelineName + "-run-",
		},
		Spec: tknv1.PipelineRunSpec{
			PipelineRef: &tknv1.PipelineRef{
				Name: pipelineName,
			},
			Params: m.GeneratePipelineRunParams(),
			Workspaces: []tknv1.WorkspaceBinding{
				{
					Name: "source",
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc,
					},
				},
				{
					Name: "dockerconfig",
					Secret: &corev1.SecretVolumeSource{
						SecretName: "aliyun-docker-config",
					},
				},
				{
					Name: "ssh",
					Secret: &corev1.SecretVolumeSource{
						SecretName: "git-ssh-secret",
					},
				},
			},
		},
	}
}

func (m *Manifest) GeneratePipelineRunParams() []tknv1.Param {
	imageTag := m.Name
	if m.Branch != "main" {
		imageTag = fmt.Sprintf("%s-%s", m.Branch, imageTag)
	}

	safeImageTag := strings.ReplaceAll(imageTag, "/", "-")
	return []tknv1.Param{
		{
			Name: "manifest-id",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: m.ID.Hex(),
			},
		},
		{
			Name: "git-url",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: m.GitRepo,
			},
		},
		{
			Name: "git-revision",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: m.Branch,
			},
		},
		{
			Name: "image-registry",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: "registry.cn-hangzhou.aliyuncs.com/devflow",
			},
		},
		{
			Name: "name",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: m.ApplicationName,
			},
		},
		{
			Name: "image-tag",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: safeImageTag,
			},
		},
		{
			Name: "manifest-name",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: m.Name,
			},
		},
	}
}
