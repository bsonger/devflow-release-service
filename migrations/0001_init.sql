CREATE TABLE IF NOT EXISTS execution_intents (
  id UUID PRIMARY KEY,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id UUID NOT NULL,
  application_id UUID NOT NULL,
  manifest_id UUID NULL,
  release_id UUID NULL,
  release_type TEXT NOT NULL DEFAULT '',
  env TEXT NOT NULL DEFAULT '',
  repo_address TEXT NOT NULL DEFAULT '',
  branch TEXT NOT NULL DEFAULT '',
  external_ref TEXT NOT NULL DEFAULT '',
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

CREATE INDEX IF NOT EXISTS idx_execution_intents_application_id
  ON execution_intents (application_id);

CREATE INDEX IF NOT EXISTS idx_execution_intents_status_kind
  ON execution_intents (status, kind);

CREATE INDEX IF NOT EXISTS idx_execution_intents_claimed_by
  ON execution_intents (claimed_by);

CREATE TABLE IF NOT EXISTS manifests (
  id UUID PRIMARY KEY,
  execution_intent_id UUID NULL,
  application_id UUID NOT NULL,
  name TEXT NOT NULL,
  branch TEXT NOT NULL,
  repo_address TEXT NOT NULL,
  commit_hash TEXT NOT NULL DEFAULT '',
  replica INTEGER NULL,
  digest TEXT NOT NULL DEFAULT '',
  type TEXT NOT NULL,
  services JSONB NOT NULL DEFAULT '[]'::jsonb,
  pipeline_id TEXT NOT NULL DEFAULT '',
  steps JSONB NOT NULL DEFAULT '[]'::jsonb,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_manifests_application_name_active
  ON manifests (application_id, name)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_manifests_pipeline_id_active
  ON manifests (pipeline_id)
  WHERE pipeline_id <> '' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_manifests_application_id
  ON manifests (application_id);

CREATE INDEX IF NOT EXISTS idx_manifests_execution_intent_id
  ON manifests (execution_intent_id);

CREATE TABLE IF NOT EXISTS releases (
  id UUID PRIMARY KEY,
  execution_intent_id UUID NULL,
  configuration_id UUID NULL,
  configuration_revision_id UUID NULL,
  application_id UUID NOT NULL,
  manifest_id UUID NOT NULL,
  env TEXT NOT NULL,
  type TEXT NOT NULL,
  steps JSONB NOT NULL DEFAULT '[]'::jsonb,
  status TEXT NOT NULL,
  external_ref TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_releases_manifest_id
  ON releases (manifest_id);

CREATE INDEX IF NOT EXISTS idx_releases_application_id
  ON releases (application_id);

CREATE INDEX IF NOT EXISTS idx_releases_configuration_id
  ON releases (configuration_id);

CREATE INDEX IF NOT EXISTS idx_releases_configuration_revision_id
  ON releases (configuration_revision_id);

CREATE INDEX IF NOT EXISTS idx_releases_execution_intent_id
  ON releases (execution_intent_id);

CREATE INDEX IF NOT EXISTS idx_releases_application_env
  ON releases (application_id, env);
