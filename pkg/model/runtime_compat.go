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
)

const (
	TraceIDAnnotation = "otel.devflow.io/trace-id"
	SpanAnnotation    = "otel.devflow.io/parent-span-id"
)

func GenerateManifestVersion(name string) string {
	t := time.Now().Format("20060102150405")
	r := rand.Intn(100)
	return fmt.Sprintf("%s%s%s", name, t, strconv.Itoa(r))
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
			PipelineRef: &tknv1.PipelineRef{Name: pipelineName},
			Params:      m.GeneratePipelineRunParams(),
			Workspaces: []tknv1.WorkspaceBinding{
				{
					Name: "source",
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc,
					},
				},
				{
					Name:   "dockerconfig",
					Secret: &corev1.SecretVolumeSource{SecretName: "aliyun-docker-config"},
				},
				{
					Name:   "ssh",
					Secret: &corev1.SecretVolumeSource{SecretName: "git-ssh-secret"},
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
			Name:  "manifest-id",
			Value: tknv1.ParamValue{Type: tknv1.ParamTypeString, StringVal: m.ID.String()},
		},
		{
			Name:  "git-url",
			Value: tknv1.ParamValue{Type: tknv1.ParamTypeString, StringVal: m.RepoAddress},
		},
		{
			Name:  "git-revision",
			Value: tknv1.ParamValue{Type: tknv1.ParamTypeString, StringVal: m.Branch},
		},
		{
			Name:  "image-registry",
			Value: tknv1.ParamValue{Type: tknv1.ParamTypeString, StringVal: "registry.cn-hangzhou.aliyuncs.com/devflow"},
		},
		{
			Name:  "name",
			Value: tknv1.ParamValue{Type: tknv1.ParamTypeString, StringVal: m.Name},
		},
		{
			Name:  "image-tag",
			Value: tknv1.ParamValue{Type: tknv1.ParamTypeString, StringVal: safeImageTag},
		},
		{
			Name:  "manifest-name",
			Value: tknv1.ParamValue{Type: tknv1.ParamTypeString, StringVal: m.Name},
		},
	}
}
