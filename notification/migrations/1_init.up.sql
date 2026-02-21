CREATE TABLE users
(
    id    INTEGER PRIMARY KEY,
    email TEXT NOT NULL,
);

CREATE INDEX IF NOT EXISTS user_idx ON users(user_id);