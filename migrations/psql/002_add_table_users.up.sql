CREATE TABLE IF NOT EXISTS users (
                                     _id           TEXT,
                                     name          TEXT,
                                     email         TEXT UNIQUE,
                                     password      BYTEA,
                                     role          TEXT CHECK (role IN ('admin', 'member')),
                                     refreshtoken  TEXT DEFAULT '',
                                     Otp_enabled   BOOLEAN DEFAULT FALSE,
                                     Otp_verified  BOOLEAN DEFAULT FALSE,
                                     Otp_secret    TEXT DEFAULT '',
                                     Otp_auth_url  TEXT DEFAULT ''
);

CREATE INDEX idx_user_email ON users (email);
alter table users owner to root;