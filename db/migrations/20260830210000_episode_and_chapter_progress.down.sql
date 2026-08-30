DROP TRIGGER IF EXISTS user_work_chapter_set_updated_at ON user_work_chapter;
DROP TRIGGER IF EXISTS user_anime_episode_set_updated_at ON user_anime_episode;
-- set_updated_at is left alone: user_anime and user_work still use it, and this
-- migration did not create it.
DROP INDEX IF EXISTS idx_user_work_chapter_user_work;
DROP INDEX IF EXISTS idx_user_work_chapter_unique;
DROP INDEX IF EXISTS idx_user_anime_episode_user_anime;
DROP INDEX IF EXISTS idx_user_anime_episode_unique;
DROP TABLE IF EXISTS user_work_chapter;
DROP TABLE IF EXISTS user_anime_episode;
