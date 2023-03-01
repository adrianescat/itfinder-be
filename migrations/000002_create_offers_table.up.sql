CREATE TABLE IF NOT EXISTS offers (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    deleted_at timestamp(0) with time zone DEFAULT null,
    active bool NOT NULL DEFAULT false,
    picture_url text,
    title text NOT NULL,
    description text NOT NULL,
    salary jsonb,
    version integer NOT NULL DEFAULT 1
);