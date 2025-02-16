CREATE TABLE IF NOT EXISTS transactions
(
    id           SERIAL PRIMARY KEY,
    from_user_id INT REFERENCES users (id),
    to_user_id   INT REFERENCES users (id),
    amount       INT       NOT NULL,
    created_at   TIMESTAMP NOT NULL
);