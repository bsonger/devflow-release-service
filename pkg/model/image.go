package model

import "github.com/google/uuid"

type Image = Manifest
type ImageTask = ManifestStep

type CreateImageRequest struct {
	ApplicationID           uuid.UUID  `json:"application_id"`
	ConfigurationRevisionID *uuid.UUID `json:"configuration_revision_id,omitempty"`
	RuntimeSpecRevisionID   *uuid.UUID `json:"runtime_spec_revision_id,omitempty"`
	Branch                  string     `json:"branch,omitempty"`
}

type PatchImageRequest = PatchManifestRequest
