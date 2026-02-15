CREATE TABLE tasks
(
    id          INTEGER PRIMARY KEY,
    name        TEXT      NOT NULL,
    description TEXT,
    starts_at   TIMESTAMP NOT NULL,
    user_id     INTEGER   NOT NULL,
    status      TEXT      NOT NULL
);

CREATE INDEX IF NOT EXISTS user_idx ON tasks(user_id);