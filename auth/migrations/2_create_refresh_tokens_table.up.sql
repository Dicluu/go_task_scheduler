CREATE TABLE IF NOT EXISTS refresh_tokens
(
    id         INTEGER PRIMARY KEY,
    user_id    INTEGER NOT NULL,
    token      varchar(255) UNIQUE,
    used       BOOL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_token ON refresh_tokens (token);