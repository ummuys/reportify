CREATE TABLE IF NOT EXISTS identity.queries (
    user_id INT PRIMARY KEY REFERENCES identity.users(user_id) ON DELETE CASCADE,
    list JSONB NOT NULL
);

COMMENT ON TABLE identity.queries IS 'История запросов пользователя';
COMMENT ON COLUMN identity.queries.user_id IS  'Уникальный идентификатор пользователя';
COMMENT ON COLUMN identity.queries.list IS 'Список запросов у пользователя';