BEGIN;

DROP INDEX IF EXISTS uq_images_application_name_active;
DROP INDEX IF EXISTS uq_manifests_application_name_active;

COMMIT;
