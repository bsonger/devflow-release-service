DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'execution_intents') THEN
    RAISE EXCEPTION 'missing table: execution_intents';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'images') THEN
    RAISE EXCEPTION 'missing table: images';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'releases') THEN
    RAISE EXCEPTION 'missing table: releases';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'manifests') THEN
    RAISE EXCEPTION 'missing table: manifests';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'images' AND column_name = 'repo_address'
  ) THEN
    RAISE EXCEPTION 'missing column: images.repo_address';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'images' AND column_name = 'runtime_spec_revision_id'
  ) THEN
    RAISE EXCEPTION 'missing column: images.runtime_spec_revision_id';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'releases' AND column_name = 'manifest_id'
  ) THEN
    RAISE EXCEPTION 'missing column: releases.manifest_id';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'manifests' AND column_name = 'artifact_repository'
  ) THEN
    RAISE EXCEPTION 'missing column: manifests.artifact_repository';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'manifests' AND column_name = 'environment_id' AND data_type = 'text'
  ) THEN
    RAISE EXCEPTION 'manifests.environment_id must be text';
  END IF;
END $$;
