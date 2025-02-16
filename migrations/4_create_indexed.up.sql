CREATE INDEX CONCURRENTLY users_username_idx ON users(username);
CREATE INDEX CONCURRENTLY transactions_from_user_idx ON transactions(from_user_id);
CREATE INDEX CONCURRENTLY transactions_to_user_idx ON transactions(to_user_id);