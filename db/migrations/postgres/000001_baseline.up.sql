-- Postgres baseline for list-service.
--
-- Generated from the production schema dumped out of PlanetScale, not from
-- this repository's MySQL migrations. The migrations describe what was
-- intended; the dump describes what is actually there, and on anime-api those
-- two disagreed -- a table its migrations create had been dropped from
-- production years ago and nothing noticed.
--
-- Worth knowing while reading this: every service shares one MySQL database
-- today. These tables sit alongside anime-api's, with a separate
-- __migrations_* tracker per service. Whether that stays true in Postgres is
-- an open decision; this file only covers the tables this service owns.

CREATE TABLE user_anime (
  id varchar(36) NOT NULL,
  user_id varchar(36) NOT NULL,
  anime_id varchar(36) NOT NULL,
  status varchar(30),
  score real DEFAULT '0',
  episodes integer DEFAULT '0',
  rewatching integer DEFAULT '0',
  rewatching_episodes integer DEFAULT '0',
  tags varchar(255),
  list_id varchar(36),
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now(),
  deleted_at timestamptz,
  PRIMARY KEY (id)
);

CREATE INDEX idx_user_anime_user_id ON user_anime (user_id);
CREATE INDEX idx_user_anime_anime_id ON user_anime (anime_id);
CREATE INDEX idx_user_anime_list_id ON user_anime (list_id);
CREATE INDEX idx_user_anime_status ON user_anime (status);
CREATE INDEX idx_user_anime_user_id_status ON user_anime (user_id, status);
CREATE INDEX idx_user_anime_created_at ON user_anime (created_at DESC);
CREATE INDEX idx_user_anime_deleted_at ON user_anime (deleted_at);
CREATE INDEX idx_user_anime_user_id_anime_id ON user_anime (user_id, anime_id);

CREATE TABLE user_list (
  id varchar(36) NOT NULL,
  user_id varchar(36) NOT NULL,
  name varchar(255) NOT NULL,
  description varchar(255),
  tags varchar(255),
  is_public boolean DEFAULT true,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now(),
  deleted_at timestamptz,
  PRIMARY KEY (id)
);

CREATE INDEX idx_user_list_user_id ON user_list (user_id);
CREATE INDEX idx_user_list_name ON user_list (name);
CREATE INDEX idx_user_list_user_id_name ON user_list (user_id, name);
CREATE INDEX idx_user_list_created_at ON user_list (created_at DESC);
CREATE INDEX idx_user_list_deleted_at ON user_list (deleted_at);

-- MySQL's ON UPDATE CURRENT_TIMESTAMP has no Postgres equivalent. Without
-- this trigger the column simply stops advancing on UPDATE -- nothing
-- errors, the value just silently goes stale.
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_anime_set_updated_at BEFORE UPDATE ON user_anime
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER user_list_set_updated_at BEFORE UPDATE ON user_list
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

