CREATE TABLE IF NOT EXISTS inventory
(
    user_id  INT REFERENCES users (id),
    item     VARCHAR(255),
    quantity INT,
    UNIQUE (user_id, item)
);