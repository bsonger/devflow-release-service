CREATE TABLE IF NOT EXISTS manifests (
  id UUID PRIMARY KEY,
  application_id UUID NOT NULL,
  environment_id UUID NOT NULL,
  image_id UUID NOT NULL,
  image_ref TEXT NOT NULL,
  services_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb,
  routes_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb,
  app_config_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
  workload_config_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
  rendered_objects JSONB NOT NULL DEFAULT '[]'::jsonb,
  rendered_yaml TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_manifests_application_id
  ON manifests (application_id);

CREATE INDEX IF NOT EXISTS idx_manifests_environment_id
  ON manifests (environment_id);

CREATE INDEX IF NOT EXISTS idx_manifests_image_id
  ON manifests (image_id);

CREATE INDEX IF NOT EXISTS idx_manifests_application_environment
  ON manifests (application_id, environment_id);
