-- Drop the unique index on refresh_tokens.user_id to support multiple concurrent sessions
-- (login from multiple browser tabs / devices without invalidating each other).
--
-- Run this manually against your production database once, or it will be executed
-- automatically on startup when ENV=production if REFRESH_TOKEN_MIGRATE=true is set.

-- Find and drop the unique index on user_id in refresh_tokens
-- The index name varies between environments; we search by column+table
DO $$
DECLARE
  idx_name TEXT;
BEGIN
  SELECT i.relname INTO idx_name
  FROM pg_index ix
  JOIN pg_class i ON i.oid = ix.indexrelid
  JOIN pg_class t ON t.oid = ix.indrelid
  JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
  WHERE t.relname = 'refresh_tokens'
    AND a.attname = 'user_id'
    AND ix.indisunique = true;

  IF idx_name IS NOT NULL THEN
    EXECUTE 'DROP INDEX IF EXISTS ' || idx_name || ' CASCADE';
    RAISE NOTICE 'Dropped unique index: %', idx_name;
  ELSE
    RAISE NOTICE 'No unique index on refresh_tokens.user_id found — already migrated.';
  END IF;
END $$;

-- Create a non-unique index to maintain query performance
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);

-- Also drop the unique constraint if it exists as a table constraint
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_user_id_key;
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS uni_refresh_tokens_user_id;