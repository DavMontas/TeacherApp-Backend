CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id           bigserial PRIMARY KEY,
    username     varchar(100) NOT NULL,
    email        citext UNIQUE NOT NULL,
    password     bytea NOT NULL,
    role         smallint,
    version      INT DEFAULT 0,
    is_active    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   timestamp(0) with time zone NOT NULL DEFAULT NOW()
);