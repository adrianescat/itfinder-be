CREATE TABLE IF NOT EXISTS profiles (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    about text NOT NULL,
    status text NOT NULL CHECK (status IN ('open', 'idle', 'close')),
    country text NOT NULL,
    state text NOT NULL,
    city text NOT NULL,
    picture_url text NOT NULL,
    website_url text NOT NULL,
    salary jsonb,
    version integer NOT NULL DEFAULT 1
);