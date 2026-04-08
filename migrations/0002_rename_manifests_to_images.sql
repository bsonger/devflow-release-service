BEGIN;

ALTER TABLE IF EXISTS manifests RENAME TO images;

ALTER INDEX IF EXISTS manifests_pkey RENAME TO images_pkey;
ALTER INDEX IF EXISTS uq_manifests_application_name_active RENAME TO uq_images_application_name_active;
ALTER INDEX IF EXISTS uq_manifests_pipeline_id_active RENAME TO uq_images_pipeline_id_active;
ALTER INDEX IF EXISTS idx_manifests_application_id RENAME TO idx_images_application_id;
ALTER INDEX IF EXISTS idx_manifests_execution_intent_id RENAME TO idx_images_execution_intent_id;

ALTER TABLE IF EXISTS releases RENAME COLUMN manifest_id TO image_id;
ALTER INDEX IF EXISTS idx_releases_manifest_id RENAME TO idx_releases_image_id;

COMMIT;
