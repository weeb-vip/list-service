-- Restores idx_user_anime_user_id. Also alone and CONCURRENTLY, same reason.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_anime_user_id ON user_anime (user_id);
