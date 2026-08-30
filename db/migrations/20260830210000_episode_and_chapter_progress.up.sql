-- Which episodes and chapters a user has actually finished.
--
-- Until now progress was a single count -- user_anime.episodes -- so the app
-- could say "9 watched" but not which nine, could not tell a rewatch from a
-- first pass, and could not show a tick against an individual episode.

-- Keyed on the episode NUMBER, not the episode row's id.
--
-- This is the whole design decision. The scraper sometimes clears an anime's
-- episodes and reinserts them, which gives every episode a new id; anything
-- pointing at those ids would be orphaned, and orphaned here means every user
-- silently loses their watch history on a re-scrape. A number survives the rows
-- being replaced, because it describes the episode rather than the record.
--
-- It also makes chapters symmetric: works have no chapter rows at all -- the
-- scraper stores a count -- so a chapter could only ever be identified by
-- number.
CREATE TABLE IF NOT EXISTS user_anime_episode
(
    id             VARCHAR(36) PRIMARY KEY,
    user_id        VARCHAR(36) NOT NULL,
    anime_id       VARCHAR(36) NOT NULL,
    episode_number INT         NOT NULL,
    -- When they watched it, which is not when the row was written: a user
    -- catching up marks six episodes in one minute, and a history that says so
    -- is useless for "what did I watch last week".
    watched_at     timestamptz DEFAULT CURRENT_TIMESTAMP,
    created_at     timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamptz DEFAULT CURRENT_TIMESTAMP,
    deleted_at     timestamptz DEFAULT NULL
);

-- One row per user per episode. Marking twice is idempotent rather than a
-- second row, and the uniqueness is enforced here instead of trusted to the
-- caller.
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_anime_episode_unique
    ON user_anime_episode (user_id, anime_id, episode_number)
    WHERE deleted_at IS NULL;

-- The read every show page performs: every episode this user has finished for
-- one anime.
CREATE INDEX IF NOT EXISTS idx_user_anime_episode_user_anime
    ON user_anime_episode (user_id, anime_id);

CREATE TABLE IF NOT EXISTS user_work_chapter
(
    id             VARCHAR(36) PRIMARY KEY,
    user_id        VARCHAR(36) NOT NULL,
    work_id        VARCHAR(36) NOT NULL,
    chapter_number INT         NOT NULL,
    read_at        timestamptz DEFAULT CURRENT_TIMESTAMP,
    created_at     timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamptz DEFAULT CURRENT_TIMESTAMP,
    deleted_at     timestamptz DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_work_chapter_unique
    ON user_work_chapter (user_id, work_id, chapter_number)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_work_chapter_user_work
    ON user_work_chapter (user_id, work_id);

-- set_updated_at comes from the user_anime migration.
CREATE TRIGGER user_anime_episode_set_updated_at BEFORE UPDATE ON user_anime_episode
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER user_work_chapter_set_updated_at BEFORE UPDATE ON user_work_chapter
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
