CREATE TABLE IF NOT EXISTS USERS
(
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    coins         INT                 NOT NULL DEFAULT 1000 CHECK ( coins >= 0 )
);