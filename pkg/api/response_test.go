package api

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewCreateResponse(t *testing.T) {
	id := primitive.NewObjectID()
	intentID := primitive.NewObjectID()

	resp := newCreateResponse(id, &intentID)

	if resp.ID != id.Hex() {
		t.Fatalf("unexpected id: got %q want %q", resp.ID, id.Hex())
	}
	if resp.ExecutionIntentID != intentID.Hex() {
		t.Fatalf("unexpected execution_intent_id: got %q want %q", resp.ExecutionIntentID, intentID.Hex())
	}
}

func TestNewCreateResponseWithoutIntent(t *testing.T) {
	id := primitive.NewObjectID()

	resp := newCreateResponse(id, nil)

	if resp.ID != id.Hex() {
		t.Fatalf("unexpected id: got %q want %q", resp.ID, id.Hex())
	}
	if resp.ExecutionIntentID != "" {
		t.Fatalf("unexpected execution_intent_id: got %q want empty", resp.ExecutionIntentID)
	}
}
