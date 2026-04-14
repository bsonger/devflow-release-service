ALTER TABLE manifests
  ALTER COLUMN environment_id TYPE TEXT
  USING environment_id::text;
