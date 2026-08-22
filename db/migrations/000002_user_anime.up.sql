CREATE TABLE IF NOT EXISTS user_anime
(
    id                  VARCHAR(36) PRIMARY KEY,
    user_id             VARCHAR(36) NOT NULL,
    anime_id            VARCHAR(36) NOT NULL,
    status              VARCHAR(30)  DEFAULT NULL,
    score               FLOAT        DEFAULT 0.0,
    episodes            INT          DEFAULT 0,
    rewatching          INT          DEFAULT 0,
    rewatching_episodes INT          DEFAULT 0,
    tags                VARCHAR(255) DEFAULT NULL,
    list_id             VARCHAR(36)  DEFAULT NULL,
    created_at          timestamptz    DEFAULT CURRENT_TIMESTAMP,
    updated_at          timestamptz    DEFAULT CURRENT_TIMESTAMP,
    deleted_at          timestamptz    DEFAULT NULL
);

-- MySQL's ON UPDATE CURRENT_TIMESTAMP has no Postgres equivalent, so the columns
-- that relied on it need a trigger to keep behaving the same way. Without it
-- updated_at silently stops advancing on UPDATE: nothing errors, the value just
-- goes stale.
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_anime_set_updated_at BEFORE UPDATE ON user_anime
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
