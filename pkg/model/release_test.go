package model

import "testing"

func TestDefaultReleaseStepsNormal(t *testing.T) {
	steps := DefaultReleaseSteps(Normal, ReleaseUpgrade)
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

func TestDefaultReleaseStepsCanary(t *testing.T) {
	steps := DefaultReleaseSteps(Canary, ReleaseUpgrade)
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

func TestDefaultReleaseStepsBlueGreenRollback(t *testing.T) {
	steps := DefaultReleaseSteps(BlueGreen, ReleaseRollback)
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

func TestDeriveReleaseStatusFromSteps(t *testing.T) {
	tests := []struct {
		name          string
		releaseAction string
		currentStatus ReleaseStatus
		steps         []ReleaseStep
		want          ReleaseStatus
	}{
		{
			name:          "pending when all pending",
			releaseAction: ReleaseUpgrade,
			steps: []ReleaseStep{
				{Name: "apply", Status: StepPending},
				{Name: "deploy", Status: StepPending},
			},
			want: ReleasePending,
		},
		{
			name:          "running when some started",
			releaseAction: ReleaseUpgrade,
			steps: []ReleaseStep{
				{Name: "apply", Status: StepSucceeded},
				{Name: "deploy", Status: StepRunning},
			},
			want: ReleaseRunning,
		},
		{
			name:          "succeeded when all succeeded",
			releaseAction: ReleaseUpgrade,
			steps: []ReleaseStep{
				{Name: "apply", Status: StepSucceeded},
				{Name: "deploy", Status: StepSucceeded},
			},
			want: ReleaseSucceeded,
		},
		{
			name:          "rolled back for rollback release",
			releaseAction: ReleaseRollback,
			steps: []ReleaseStep{
				{Name: "apply rollback", Status: StepSucceeded},
				{Name: "deploy ready", Status: StepSucceeded},
			},
			want: ReleaseRolledBack,
		},
		{
			name:          "failed when one step failed",
			releaseAction: ReleaseUpgrade,
			steps: []ReleaseStep{
				{Name: "apply", Status: StepSucceeded},
				{Name: "deploy", Status: StepFailed},
			},
			want: ReleaseFailed,
		},
		{
			name:          "preserve terminal sync failed",
			releaseAction: ReleaseUpgrade,
			currentStatus: ReleaseSyncFailed,
			steps: []ReleaseStep{
				{Name: "apply", Status: StepSucceeded},
				{Name: "deploy", Status: StepSucceeded},
			},
			want: ReleaseSyncFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveReleaseStatusFromSteps(tt.releaseAction, tt.currentStatus, tt.steps)
			if got != tt.want {
				t.Fatalf("unexpected status: got %q want %q", got, tt.want)
			}
		})
	}
}
