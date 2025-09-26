-- Add composite index for user_id and anime_id optimization
CREATE INDEX idx_user_anime_user_id_anime_id ON user_anime(user_id, anime_id);