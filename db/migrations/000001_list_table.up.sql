CREATE TABLE IF NOT EXISTS user_list
(
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36)  NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description VARCHAR(255) DEFAULT NULL,
    tags        VARCHAR(255) DEFAULT NULL,
    is_public   BOOLEAN      DEFAULT TRUE,
    created_at  timestamptz    DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamptz    DEFAULT CURRENT_TIMESTAMP,
    deleted_at  timestamptz    DEFAULT NULL
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

CREATE TRIGGER user_list_set_updated_at BEFORE UPDATE ON user_list
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
