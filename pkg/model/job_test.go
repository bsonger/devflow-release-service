package model

import "testing"

func TestDefaultJobStepsNormal(t *testing.T) {
	steps := DefaultJobSteps(Normal, JobUpgrade)
	if len(steps) != 2 {
		t.Fatalf("unexpected step count: got %d want 2", len(steps))
	}
	if steps[0].Name != "apply manifests" {
		t.Fatalf("unexpected first step: got %q want %q", steps[0].Name, "apply manifests")
	}
	if steps[1].Name != "deploy ready" {
		t.Fatalf("unexpected second step: got %q want %q", steps[1].Name, "deploy ready")
	}
}

func TestDefaultJobStepsCanary(t *testing.T) {
	steps := DefaultJobSteps(Canary, JobUpgrade)
	if len(steps) != 5 {
		t.Fatalf("unexpected step count: got %d want 5", len(steps))
	}
	if steps[1].Name != "canary 10% traffic" {
		t.Fatalf("unexpected canary step: got %q want %q", steps[1].Name, "canary 10% traffic")
	}
	if steps[4].Name != "canary 100% traffic" {
		t.Fatalf("unexpected last canary step: got %q want %q", steps[4].Name, "canary 100% traffic")
	}
}

func TestDefaultJobStepsBlueGreenRollback(t *testing.T) {
	steps := DefaultJobSteps(BlueGreen, JobRollback)
	if len(steps) != 3 {
		t.Fatalf("unexpected step count: got %d want 3", len(steps))
	}
	if steps[0].Name != "apply rollback manifests" {
		t.Fatalf("unexpected first step: got %q want %q", steps[0].Name, "apply rollback manifests")
	}
	if steps[1].Name != "green ready" {
		t.Fatalf("unexpected green step: got %q want %q", steps[1].Name, "green ready")
	}
	if steps[2].Name != "switch traffic" {
		t.Fatalf("unexpected traffic step: got %q want %q", steps[2].Name, "switch traffic")
	}
}

func TestDeriveJobStatusFromSteps(t *testing.T) {
	tests := []struct {
		name          string
		jobType       string
		currentStatus JobStatus
		steps         []JobStep
		want          JobStatus
	}{
		{
			name:    "pending when all pending",
			jobType: JobUpgrade,
			steps: []JobStep{
				{Name: "apply", Status: StepPending},
				{Name: "deploy", Status: StepPending},
			},
			want: JobPending,
		},
		{
			name:    "running when some started",
			jobType: JobUpgrade,
			steps: []JobStep{
				{Name: "apply", Status: StepSucceeded},
				{Name: "deploy", Status: StepRunning},
			},
			want: JobRunning,
		},
		{
			name:    "succeeded when all succeeded",
			jobType: JobUpgrade,
			steps: []JobStep{
				{Name: "apply", Status: StepSucceeded},
				{Name: "deploy", Status: StepSucceeded},
			},
			want: JobSucceeded,
		},
		{
			name:    "rolled back for rollback job",
			jobType: JobRollback,
			steps: []JobStep{
				{Name: "apply rollback", Status: StepSucceeded},
				{Name: "deploy ready", Status: StepSucceeded},
			},
			want: JobRolledBack,
		},
		{
			name:    "failed when one step failed",
			jobType: JobUpgrade,
			steps: []JobStep{
				{Name: "apply", Status: StepSucceeded},
				{Name: "deploy", Status: StepFailed},
			},
			want: JobFailed,
		},
		{
			name:          "preserve terminal sync failed",
			jobType:       JobUpgrade,
			currentStatus: JobSyncFailed,
			steps: []JobStep{
				{Name: "apply", Status: StepSucceeded},
				{Name: "deploy", Status: StepSucceeded},
			},
			want: JobSyncFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveJobStatusFromSteps(tt.jobType, tt.currentStatus, tt.steps)
			if got != tt.want {
				t.Fatalf("unexpected status: got %q want %q", got, tt.want)
			}
		})
	}
}
