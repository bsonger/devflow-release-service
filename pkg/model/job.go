package model

import (
	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"go.mongodb.org/mongo-driver/bson/primitive"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type JobStatus string

const (
	JobPending     JobStatus = "Pending"
	JobRunning     JobStatus = "Running"
	JobSucceeded   JobStatus = "Succeeded"
	JobFailed      JobStatus = "Failed"
	JobRollingBack JobStatus = "RollingBack"
	JobRolledBack  JobStatus = "RolledBack"
	JobSyncing     JobStatus = "Syncing"
	JobSyncFailed  JobStatus = "SyncFailed"

	JobInstall  string = "Install"
	JobUpgrade  string = "Upgrade"
	JobRollback string = "Rollback"

	JobIDLabel = "devflow.io/job-id"

	defaultArgoProject = "app"
)

type Job struct {
	BaseModel `bson:",inline"`

	ExecutionIntentID *primitive.ObjectID `bson:"execution_intent_id,omitempty" json:"execution_intent_id,omitempty"`
	ApplicationId     primitive.ObjectID  `bson:"application_id" json:"application_id"`
	ApplicationName   string              `bson:"application_name" json:"application_name"`
	ProjectName       string              `bson:"project_name" json:"project_name"`
	ManifestID        primitive.ObjectID  `bson:"manifest_id" json:"manifest_id"`
	ManifestName      string              `bson:"manifest_name" json:"manifest_name"`
	Type              string              `bson:"type" json:"type"`
	Env               string              `bson:"env" json:"env"`
	Status            JobStatus           `bson:"status" json:"status"`
	Steps             []JobStep           `bson:"steps,omitempty" json:"steps,omitempty"`
}

func (j *Job) CollectionName() string { return "job" }

type JobStep struct {
	Name      string     `bson:"name" json:"name"`
	Progress  int32      `bson:"progress" json:"progress"`
	Status    StepStatus `bson:"status" json:"status"`
	Message   string     `bson:"message,omitempty" json:"message,omitempty"`
	StartTime *time.Time `bson:"start_time,omitempty" json:"start_time,omitempty"`
	EndTime   *time.Time `bson:"end_time,omitempty" json:"end_time,omitempty"`
}

func DeriveJobStatusFromSteps(jobType string, currentStatus JobStatus, steps []JobStep) JobStatus {
	switch currentStatus {
	case JobSucceeded, JobFailed, JobRolledBack, JobSyncFailed:
		return currentStatus
	}

	if len(steps) == 0 {
		if currentStatus == "" {
			return JobPending
		}
		return currentStatus
	}

	allSucceeded := true
	anyFailed := false
	anyStarted := false

	for _, step := range steps {
		switch step.Status {
		case StepFailed:
			anyFailed = true
			allSucceeded = false
		case StepSucceeded:
			anyStarted = true
		case StepRunning:
			anyStarted = true
			allSucceeded = false
		default:
			allSucceeded = false
		}
	}

	if anyFailed {
		return JobFailed
	}
	if allSucceeded {
		if jobType == JobRollback {
			return JobRolledBack
		}
		return JobSucceeded
	}
	if anyStarted {
		return JobRunning
	}
	if currentStatus == "" {
		return JobPending
	}
	return currentStatus
}

func DefaultJobSteps(releaseType ReleaseType, jobType string) []JobStep {
	applyStepName := "apply manifests"
	switch jobType {
	case JobRollback:
		applyStepName = "apply rollback manifests"
	case JobInstall:
		applyStepName = "apply install manifests"
	}

	stepNames := []string{applyStepName}
	switch releaseType {
	case Canary:
		stepNames = append(stepNames,
			"canary 10% traffic",
			"canary 30% traffic",
			"canary 60% traffic",
			"canary 100% traffic",
		)
	case BlueGreen:
		stepNames = append(stepNames,
			"green ready",
			"switch traffic",
		)
	default:
		stepNames = append(stepNames, "deploy ready")
	}

	steps := make([]JobStep, 0, len(stepNames))
	for _, name := range stepNames {
		steps = append(steps, JobStep{
			Name:     name,
			Progress: 0,
			Status:   StepPending,
		})
	}

	return steps
}

func (j *Job) GenerateApplication() *appv1.Application {
	manifestID := j.ManifestID.Hex()
	jobID := j.ID.Hex()

	return &appv1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: j.ApplicationName,
		},
		Spec: appv1.ApplicationSpec{
			Project: defaultArgoProject,
			Source: &appv1.ApplicationSource{
				RepoURL: manifestRepo.Address,
				Path:    "./",
				Plugin: &appv1.ApplicationSourcePlugin{
					Name: "plugin",
					Parameters: []appv1.ApplicationSourcePluginParameter{
						{
							Name:    "env",
							String_: &j.Env,
						},
						{
							Name:    "manifest-id",
							String_: &manifestID,
						},
						{
							Name:    "job-id",
							String_: &jobID,
						},
					},
				},
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: j.ProjectName,
			},
		},
	}
}
