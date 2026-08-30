-- Tracking for works: the manga, light novels and novels behind /manga/<slug>.
--
-- A separate table rather than a media_type column on user_anime. The two rows
-- carry different progress -- episodes against chapters and volumes -- and
-- widening user_anime would leave every anime row with two null columns and
-- every work row with a null episodes column, on the one table every anime card
-- on the site already reads.
CREATE TABLE IF NOT EXISTS user_work
(
    id                  VARCHAR(36) PRIMARY KEY,
    user_id             VARCHAR(36) NOT NULL,
    work_id             VARCHAR(36) NOT NULL,
    status              VARCHAR(30)  DEFAULT NULL,
    score               FLOAT        DEFAULT 0.0,
    -- Both, because a reader tracks whichever the release is published in and
    -- MyAnimeList reports both counts.
    chapters            INT          DEFAULT 0,
    volumes             INT          DEFAULT 0,
    tags                VARCHAR(255) DEFAULT NULL,
    list_id             VARCHAR(36)  DEFAULT NULL,
    created_at          timestamptz  DEFAULT CURRENT_TIMESTAMP,
    updated_at          timestamptz  DEFAULT CURRENT_TIMESTAMP,
    deleted_at          timestamptz  DEFAULT NULL
);

-- set_updated_at already exists, created by the user_anime migration. Reused
-- rather than redefined: two CREATE OR REPLACE of the same function in one
-- schema is a race waiting to happen when migrations are reordered.
CREATE TRIGGER user_work_set_updated_at BEFORE UPDATE ON user_work
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- The lookup every read performs: "does this user have a row for this work".
-- The dataloader batches it as user_id + work_id IN (...), which uses the same
-- index.
CREATE INDEX IF NOT EXISTS idx_user_work_user_id_work_id ON user_work(user_id, work_id);

-- Listing a user's shelf, filtered by status and ordered by recency.
CREATE INDEX IF NOT EXISTS idx_user_work_user_id_status ON user_work(user_id, status);
