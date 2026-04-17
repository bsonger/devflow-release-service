BEGIN;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'manifests'
  ) AND NOT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'images'
  ) THEN
    ALTER TABLE manifests RENAME TO images;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'manifests_pkey')
     AND NOT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'images_pkey') THEN
    ALTER INDEX manifests_pkey RENAME TO images_pkey;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'uq_manifests_application_name_active')
     AND NOT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'uq_images_application_name_active') THEN
    ALTER INDEX uq_manifests_application_name_active RENAME TO uq_images_application_name_active;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'uq_manifests_pipeline_id_active')
     AND NOT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'uq_images_pipeline_id_active') THEN
    ALTER INDEX uq_manifests_pipeline_id_active RENAME TO uq_images_pipeline_id_active;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'idx_manifests_application_id')
     AND NOT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'idx_images_application_id') THEN
    ALTER INDEX idx_manifests_application_id RENAME TO idx_images_application_id;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'idx_manifests_execution_intent_id')
     AND NOT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'idx_images_execution_intent_id') THEN
    ALTER INDEX idx_manifests_execution_intent_id RENAME TO idx_images_execution_intent_id;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'releases' AND column_name = 'manifest_id'
  ) AND NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'releases' AND column_name = 'image_id'
  ) THEN
    ALTER TABLE releases RENAME COLUMN manifest_id TO image_id;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'idx_releases_manifest_id')
     AND NOT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'idx_releases_image_id') THEN
    ALTER INDEX idx_releases_manifest_id RENAME TO idx_releases_image_id;
  END IF;
END $$;

COMMIT;
