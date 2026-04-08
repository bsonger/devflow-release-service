CREATE TABLE IF NOT EXISTS execution_intents (
  id UUID PRIMARY KEY,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id UUID NOT NULL,
  trace_id TEXT NOT NULL DEFAULT '',
  message TEXT NOT NULL DEFAULT '',
  last_error TEXT NOT NULL DEFAULT '',
  claimed_by TEXT NOT NULL DEFAULT '',
  claimed_at TIMESTAMPTZ NULL,
  lease_expires_at TIMESTAMPTZ NULL,
  attempt_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_execution_intents_resource
  ON execution_intents (resource_type, resource_id);

CREATE INDEX IF NOT EXISTS idx_execution_intents_status_kind
  ON execution_intents (status, kind);

CREATE INDEX IF NOT EXISTS idx_execution_intents_claimed_by
  ON execution_intents (claimed_by);

CREATE TABLE IF NOT EXISTS images (
  id UUID PRIMARY KEY,
  execution_intent_id UUID NULL,
  application_id UUID NOT NULL,
  configuration_revision_id UUID NULL,
  runtime_spec_revision_id UUID NULL,
  name TEXT NOT NULL,
  branch TEXT NOT NULL,
  repo_address TEXT NOT NULL,
  commit_hash TEXT NOT NULL DEFAULT '',
  digest TEXT NOT NULL DEFAULT '',
  pipeline_id TEXT NOT NULL DEFAULT '',
  steps JSONB NOT NULL DEFAULT '[]'::jsonb,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_images_application_name_active
  ON images (application_id, name)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_images_pipeline_id_active
  ON images (pipeline_id)
  WHERE pipeline_id <> '' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_images_application_id
  ON images (application_id);

CREATE INDEX IF NOT EXISTS idx_images_execution_intent_id
  ON images (execution_intent_id);

CREATE TABLE IF NOT EXISTS releases (
  id UUID PRIMARY KEY,
  execution_intent_id UUID NULL,
  application_id UUID NOT NULL,
  image_id UUID NOT NULL,
  env TEXT NOT NULL,
  type TEXT NOT NULL,
  steps JSONB NOT NULL DEFAULT '[]'::jsonb,
  status TEXT NOT NULL,
  external_ref TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_releases_image_id
  ON releases (image_id);

CREATE INDEX IF NOT EXISTS idx_releases_application_id
  ON releases (application_id);

CREATE INDEX IF NOT EXISTS idx_releases_execution_intent_id
  ON releases (execution_intent_id);

CREATE INDEX IF NOT EXISTS idx_releases_application_env
  ON releases (application_id, env);
