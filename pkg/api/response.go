package api

import "go.mongodb.org/mongo-driver/bson/primitive"

type CreateResponse struct {
	ID                string `json:"id"`
	ExecutionIntentID string `json:"execution_intent_id,omitempty"`
}

func newCreateResponse(id primitive.ObjectID, intentID *primitive.ObjectID) CreateResponse {
	resp := CreateResponse{
		ID: id.Hex(),
	}
	if intentID != nil && !intentID.IsZero() {
		resp.ExecutionIntentID = intentID.Hex()
	}
	return resp
}
