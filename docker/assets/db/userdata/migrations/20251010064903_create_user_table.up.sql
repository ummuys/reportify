
CREATE TABLE IF NOT EXISTS users.data (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


COMMENT ON TABLE users.data IS 'Таблица для хранения данных пользователей';
COMMENT ON COLUMN users.data.id IS 'Уникальный идентификатор пользователя';
COMMENT ON COLUMN users.data.username IS 'Имя пользователя (логин)';
COMMENT ON COLUMN users.data.password IS 'Хэшированный пароль пользователя';
COMMENT ON COLUMN users.data.created_at IS 'Дата и время создания пользователя';