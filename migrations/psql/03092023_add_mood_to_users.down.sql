BEGIN;

ALTER TABLE users
DROP COLUMN IF EXISTS refreshtoken;

COMMIT;