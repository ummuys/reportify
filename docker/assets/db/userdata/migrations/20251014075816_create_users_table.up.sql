
CREATE TABLE IF NOT EXISTS identity.users (
    user_id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


COMMENT ON TABLE identity.users IS 'Таблица для хранения данных пользователей';
COMMENT ON COLUMN identity.users.user_id IS 'Уникальный идентификатор пользователя';
COMMENT ON COLUMN identity.users.username IS 'Имя пользователя (логин)';
COMMENT ON COLUMN identity.users.password IS 'Хэшированный пароль пользователя';
COMMENT ON COLUMN identity.users.created_at IS 'Дата и время создания пользователя';