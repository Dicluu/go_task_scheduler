CREATE TABLE users
(
    id      INTEGER PRIMARY KEY,
    email   TEXT NOT NULL,
    user_id INTEGER
);

CREATE INDEX IF NOT EXISTS user_idx ON users(user_id);