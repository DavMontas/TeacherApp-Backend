ALTER TABLE 
    user_invitations
ADD
    COLUMN expiration TIMESTAMP(0) WITH TIME ZONE NOT NULL;