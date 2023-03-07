CREATE TABLE IF NOT EXISTS offers_applicants (
    profile_id bigint NOT NULL REFERENCES profiles ON DELETE CASCADE,
    offer_id bigint NOT NULL REFERENCES offers ON DELETE CASCADE,
    PRIMARY KEY (profile_id, offer_id)
);

ALTER TABLE offers_applicants ADD UNIQUE (profile_id, offer_id)