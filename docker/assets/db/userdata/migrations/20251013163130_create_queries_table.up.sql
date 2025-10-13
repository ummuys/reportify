CREATE TABLE IF NOT EXISTS users.user_queries (
    username TEXT PRIMARY KEY,
    queries TEXT[] NOT NULL DEFAULT '{}'
);