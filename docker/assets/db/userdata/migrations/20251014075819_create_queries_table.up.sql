CREATE TABLE IF NOT EXISTS identity.queries (
    user_id INT PRIMARY KEY REFERENCES identity.users(user_id) ON DELETE CASCADE,
    query text,
    report_name text,
    report_comm text,
    report_created_at TIMESTAMP
);

COMMENT ON TABLE identity.queries IS 'История запросов пользователя';
COMMENT ON COLUMN identity.queries.user_id IS 'Ссылка на пользователя в identity.users';
COMMENT ON COLUMN identity.queries.query IS 'SQL-запрос пользователя';
COMMENT ON COLUMN identity.queries.report_name IS 'Название отчета';
COMMENT ON COLUMN identity.queries.report_comm IS 'Комментарий к отчету';
COMMENT ON COLUMN identity.queries.report_created_at IS 'Время создания отчета';