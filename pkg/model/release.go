package model

import (
	"time"

	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"go.mongodb.org/mongo-driver/bson/primitive"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReleaseStatus string

const (
	ReleasePending     ReleaseStatus = "Pending"
	ReleaseRunning     ReleaseStatus = "Running"
	ReleaseSucceeded   ReleaseStatus = "Succeeded"
	ReleaseFailed      ReleaseStatus = "Failed"
	ReleaseRollingBack ReleaseStatus = "RollingBack"
	ReleaseRolledBack  ReleaseStatus = "RolledBack"
	ReleaseSyncing     ReleaseStatus = "Syncing"
	ReleaseSyncFailed  ReleaseStatus = "SyncFailed"

	ReleaseInstall  string = "Install"
	ReleaseUpgrade  string = "Upgrade"
	ReleaseRollback string = "Rollback"

	ReleaseIDLabel = "devflow.io/release-id"

	defaultArgoProject = "app"
)

type Release struct {
	BaseModel `bson:",inline"`

	ExecutionIntentID *primitive.ObjectID `bson:"execution_intent_id,omitempty" json:"execution_intent_id,omitempty"`
	ApplicationId     primitive.ObjectID  `bson:"application_id" json:"application_id"`
	ApplicationName   string              `bson:"application_name" json:"application_name"`
	ProjectName       string              `bson:"project_name" json:"project_name"`
	ManifestID        primitive.ObjectID  `bson:"manifest_id" json:"manifest_id"`
	ManifestName      string              `bson:"manifest_name" json:"manifest_name"`
	Type              string              `bson:"type" json:"type"`
	Env               string              `bson:"env" json:"env"`
	Status            ReleaseStatus       `bson:"status" json:"status"`
	Steps             []ReleaseStep       `bson:"steps,omitempty" json:"steps,omitempty"`
}

func (r *Release) CollectionName() string { return "release" }

type ReleaseStep struct {
	Name      string     `bson:"name" json:"name"`
	Progress  int32      `bson:"progress" json:"progress"`
	Status    StepStatus `bson:"status" json:"status"`
	Message   string     `bson:"message,omitempty" json:"message,omitempty"`
	StartTime *time.Time `bson:"start_time,omitempty" json:"start_time,omitempty"`
	EndTime   *time.Time `bson:"end_time,omitempty" json:"end_time,omitempty"`
}

func DeriveReleaseStatusFromSteps(releaseAction string, currentStatus ReleaseStatus, steps []ReleaseStep) ReleaseStatus {
	switch currentStatus {
	case ReleaseSucceeded, ReleaseFailed, ReleaseRolledBack, ReleaseSyncFailed:
		return currentStatus
	}

	if len(steps) == 0 {
		if currentStatus == "" {
			return ReleasePending
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
		return ReleaseFailed
	}
	if allSucceeded {
		if releaseAction == ReleaseRollback {
			return ReleaseRolledBack
		}
		return ReleaseSucceeded
	}
	if anyStarted {
		return ReleaseRunning
	}
	if currentStatus == "" {
		return ReleasePending
	}
	return currentStatus
}

func DefaultReleaseSteps(strategy ReleaseType, releaseAction string) []ReleaseStep {
	applyStepName := "apply manifests"
	switch releaseAction {
	case ReleaseRollback:
		applyStepName = "apply rollback manifests"
	case ReleaseInstall:
		applyStepName = "apply install manifests"
	}

	stepNames := []string{applyStepName}
	switch strategy {
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

	steps := make([]ReleaseStep, 0, len(stepNames))
	for _, name := range stepNames {
		steps = append(steps, ReleaseStep{
			Name:     name,
			Progress: 0,
			Status:   StepPending,
		})
	}

	return steps
}

func (r *Release) GenerateApplication() *appv1.Application {
	manifestID := r.ManifestID.Hex()
	releaseID := r.ID.Hex()

	return &appv1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: r.ApplicationName,
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
							String_: &r.Env,
						},
						{
							Name:    "manifest-id",
							String_: &manifestID,
						},
						{
							Name:    "release-id",
							String_: &releaseID,
						},
					},
				},
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: r.ProjectName,
			},
		},
	}
}
