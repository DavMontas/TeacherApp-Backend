CREATE TABLE IF NOT EXISTS bank_accounts (
    id                   bigserial PRIMARY KEY,
    bank_name            varchar(255) NOT NULL,
    bank_account_number  varchar(255) NOT NULL,
    user_profile_id          bigint NOT NULL,
    created_at           timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    CONSTRAINT           fk_profile FOREIGN KEY (user_profile_id) REFERENCES user_profiles(id)
);