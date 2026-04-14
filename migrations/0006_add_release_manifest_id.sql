ALTER TABLE releases
  ADD COLUMN IF NOT EXISTS manifest_id UUID;

UPDATE releases
SET manifest_id = '00000000-0000-0000-0000-000000000000'
WHERE manifest_id IS NULL;

ALTER TABLE releases
  ALTER COLUMN manifest_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_releases_manifest_id
  ON releases (manifest_id);
