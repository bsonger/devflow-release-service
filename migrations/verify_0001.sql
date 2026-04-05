DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'execution_intents') THEN
    RAISE EXCEPTION 'missing table: execution_intents';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'manifests') THEN
    RAISE EXCEPTION 'missing table: manifests';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'releases') THEN
    RAISE EXCEPTION 'missing table: releases';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'manifests' AND column_name = 'repo_address'
  ) THEN
    RAISE EXCEPTION 'missing column: manifests.repo_address';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'manifests' AND column_name = 'runtime_spec_revision_id'
  ) THEN
    RAISE EXCEPTION 'missing column: manifests.runtime_spec_revision_id';
  END IF;
END $$;
