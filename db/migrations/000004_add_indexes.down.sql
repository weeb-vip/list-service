-- Drop indexes for user_anime table
DROP INDEX idx_user_anime_deleted_at ON user_anime;
DROP INDEX idx_user_anime_created_at ON user_anime;
DROP INDEX idx_user_anime_user_id_status ON user_anime;
DROP INDEX idx_user_anime_status ON user_anime;
DROP INDEX idx_user_anime_list_id ON user_anime;
DROP INDEX idx_user_anime_anime_id ON user_anime;
DROP INDEX idx_user_anime_user_id ON user_anime;

-- Drop indexes for user_list table  
DROP INDEX idx_user_list_deleted_at ON user_list;
DROP INDEX idx_user_list_created_at ON user_list;
DROP INDEX idx_user_list_user_id_name ON user_list;
DROP INDEX idx_user_list_name ON user_list;
DROP INDEX idx_user_list_user_id ON user_list;