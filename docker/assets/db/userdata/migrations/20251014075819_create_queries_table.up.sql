CREATE TABLE IF NOT EXISTS identity.queries (
    user_id INT PRIMARY KEY REFERENCES identity.users(user_id) ON DELETE CASCADE,
    list TEXT[] NOT NULL DEFAULT '{}'
);

COMMENT ON TABLE identity.queries IS 'История запросов пользователя';
COMMENT ON COLUMN identity.queries.user_id IS 'Ссылка на пользователя в identity.users';
COMMENT ON COLUMN identity.queries.list IS 'Список SQL-запросов пользователя';