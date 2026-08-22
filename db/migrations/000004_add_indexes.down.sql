-- Drop indexes for user_anime table
DROP INDEX IF EXISTS idx_user_anime_deleted_at;
DROP INDEX IF EXISTS idx_user_anime_created_at;
DROP INDEX IF EXISTS idx_user_anime_user_id_status;
DROP INDEX IF EXISTS idx_user_anime_status;
DROP INDEX IF EXISTS idx_user_anime_list_id;
DROP INDEX IF EXISTS idx_user_anime_anime_id;
DROP INDEX IF EXISTS idx_user_anime_user_id;

-- Drop indexes for user_list table  
DROP INDEX IF EXISTS idx_user_list_deleted_at;
DROP INDEX IF EXISTS idx_user_list_created_at;
DROP INDEX IF EXISTS idx_user_list_user_id_name;
DROP INDEX IF EXISTS idx_user_list_name;
DROP INDEX IF EXISTS idx_user_list_user_id;