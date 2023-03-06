CREATE TABLE IF NOT EXISTS profile_bookmarks (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    profile_id bigint NOT NULL REFERENCES profiles ON DELETE CASCADE,
    PRIMARY KEY (user_id, profile_id)
);