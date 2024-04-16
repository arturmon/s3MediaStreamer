-- Dropping the index before dropping the table to ensure a clean slate
DROP INDEX IF EXISTS idx_user_email;

-- Dropping the tracks table if it exists to avoid errors in case of re-creation
DROP TABLE IF EXISTS users;