-- Add indexes for user_list table
CREATE INDEX idx_user_list_user_id ON user_list(user_id);
CREATE INDEX idx_user_list_name ON user_list(name);
CREATE INDEX idx_user_list_user_id_name ON user_list(user_id, name);
CREATE INDEX idx_user_list_created_at ON user_list(created_at DESC);
CREATE INDEX idx_user_list_deleted_at ON user_list(deleted_at);

-- Add indexes for user_anime table
CREATE INDEX idx_user_anime_user_id ON user_anime(user_id);
CREATE INDEX idx_user_anime_anime_id ON user_anime(anime_id);
CREATE INDEX idx_user_anime_list_id ON user_anime(list_id);
CREATE INDEX idx_user_anime_status ON user_anime(status);
CREATE INDEX idx_user_anime_user_id_status ON user_anime(user_id, status);
CREATE INDEX idx_user_anime_created_at ON user_anime(created_at DESC);
CREATE INDEX idx_user_anime_deleted_at ON user_anime(deleted_at);