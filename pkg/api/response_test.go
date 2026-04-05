package api

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewCreateResponse(t *testing.T) {
	id := uuid.New()
	intentID := uuid.New()

	resp := newCreateResponse(id, &intentID)

	if resp.ID != id.String() {
		t.Fatalf("unexpected id: got %q want %q", resp.ID, id.String())
	}
	if resp.ExecutionIntentID != intentID.String() {
		t.Fatalf("unexpected execution_intent_id: got %q want %q", resp.ExecutionIntentID, intentID.String())
	}
}

func TestNewCreateResponseWithoutIntent(t *testing.T) {
	id := uuid.New()

	resp := newCreateResponse(id, nil)

	if resp.ID != id.String() {
		t.Fatalf("unexpected id: got %q want %q", resp.ID, id.String())
	}
	if resp.ExecutionIntentID != "" {
		t.Fatalf("unexpected execution_intent_id: got %q want empty", resp.ExecutionIntentID)
	}
}
