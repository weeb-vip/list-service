DROP TRIGGER IF EXISTS user_work_set_updated_at ON user_work;
-- set_updated_at is left alone: user_anime's trigger still uses it, and this
-- migration did not create it.
DROP INDEX IF EXISTS idx_user_work_user_id_status;
DROP INDEX IF EXISTS idx_user_work_user_id_work_id;
DROP TABLE IF EXISTS user_work;
