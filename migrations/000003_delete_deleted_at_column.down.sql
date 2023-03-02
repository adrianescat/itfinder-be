ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at timestamp(0) with time zone DEFAULT null;
ALTER TABLE offers ADD COLUMN IF NOT EXISTS deleted_at timestamp(0) with time zone DEFAULT null;