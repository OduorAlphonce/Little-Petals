-- Add updated_at columns to the users and posts tables.
--
-- Idempotent: ADD COLUMN IF NOT EXISTS makes this safe to re-run on databases
-- where the column already exists, and safe to apply on top of an existing
-- 001_create_users / 002_create_posts schema.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP;
