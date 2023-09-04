BEGIN;

ALTER TABLE users
    ADD COLUMN refreshtoken TEXT;

COMMIT;