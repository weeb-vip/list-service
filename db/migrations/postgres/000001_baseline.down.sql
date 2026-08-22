-- Reverse of the baseline.
DROP TABLE IF EXISTS user_anime;
DROP TABLE IF EXISTS user_list;

-- Dropped explicitly: the function is schema-level rather than owned by any
-- table, so it survives the drops above and would collide on a re-run.
DROP FUNCTION IF EXISTS set_updated_at();
