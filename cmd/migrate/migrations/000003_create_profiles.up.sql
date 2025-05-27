CREATE TABLE IF NOT EXISTS user_profiles (
    id               bigserial PRIMARY KEY,
    identification   varchar(255) NULL,
    first_name       varchar(255) NULL,
    last_name        varchar(255) NULL,
    created_at       timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id          bigint NOT NULL,
    CONSTRAINT       fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);