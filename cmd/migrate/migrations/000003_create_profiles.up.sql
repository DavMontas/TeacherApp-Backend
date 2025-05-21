CREATE TABLE IF NOT EXISTS profiles (
    id               bigserial PRIMARY KEY,
    identification   varchar(255),
    first_name       varchar(255) NOT NULL,
    last_name        varchar(255) NOT NULL,
    created_at       timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id          bigint NOT NULL,
    CONSTRAINT       fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);