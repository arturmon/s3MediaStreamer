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

-- Create an index on the email column
CREATE INDEX idx_user_email ON users (email);

-- Alter table owner
ALTER TABLE users OWNER TO root;

-- Add comments on columns
COMMENT ON COLUMN users._id IS 'Unique identifier for the user';
COMMENT ON COLUMN users.name IS 'Name of the user';
COMMENT ON COLUMN users.email IS 'Email address of the user';
COMMENT ON COLUMN users.password IS 'Hashed password of the user';
COMMENT ON COLUMN users.role IS 'Role of the user (admin or member)';
COMMENT ON COLUMN users.refreshtoken IS 'Refresh token for the user';
COMMENT ON COLUMN users.Otp_enabled IS 'Flag indicating whether OTP is enabled for the user';
COMMENT ON COLUMN users.Otp_verified IS 'Flag indicating whether OTP is verified for the user';
COMMENT ON COLUMN users.Otp_secret IS 'Secret key for OTP authentication';
COMMENT ON COLUMN users.Otp_auth_url IS 'URL for OTP authentication';