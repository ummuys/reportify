
CREATE SCHEMA IF NOT EXISTS templates;
COMMENT ON SCHEMA templates IS 'Схема с шаблонными таблицами';

CREATE TABLE IF NOT EXISTS templates.users (
    user_id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    full_name TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

COMMENT ON TABLE templates.users IS 'Пример таблицы пользователей';
COMMENT ON COLUMN templates.users.user_id IS 'Уникальный идентификатор пользователя';
COMMENT ON COLUMN templates.users.username IS 'Имя пользователя (уникальное)';
COMMENT ON COLUMN templates.users.email IS 'Email пользователя';
COMMENT ON COLUMN templates.users.full_name IS 'Полное имя';
COMMENT ON COLUMN templates.users.created_at IS 'Дата регистрации';
COMMENT ON COLUMN templates.users.is_active IS 'Флаг активности пользователя';


CREATE TABLE IF NOT EXISTS templates.orders (
    order_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES templates.users(user_id) ON DELETE CASCADE,
    order_date TIMESTAMP DEFAULT NOW(),
    total_amount NUMERIC(10,2) NOT NULL,
    status TEXT DEFAULT 'new',
    shipping_address TEXT,
    note TEXT
);

COMMENT ON TABLE templates.orders IS 'Пример таблицы заказов пользователей';
COMMENT ON COLUMN templates.orders.order_id IS 'ID заказа';
COMMENT ON COLUMN templates.orders.user_id IS 'Ссылка на пользователя';
COMMENT ON COLUMN templates.orders.order_date IS 'Дата создания заказа';
COMMENT ON COLUMN templates.orders.total_amount IS 'Общая сумма заказа';
COMMENT ON COLUMN templates.orders.status IS 'Статус заказа';
COMMENT ON COLUMN templates.orders.shipping_address IS 'Адрес доставки';
COMMENT ON COLUMN templates.orders.note IS 'Примечание к заказу';

CREATE TABLE IF NOT EXISTS templates.logs (
    log_id SERIAL PRIMARY KEY,
    event_time TIMESTAMP DEFAULT NOW(),
    level TEXT DEFAULT 'INFO',
    message TEXT,
    metadata JSONB
);

COMMENT ON TABLE templates.logs IS 'Пример таблицы логов системы';
COMMENT ON COLUMN templates.logs.log_id IS 'Уникальный ID записи';
COMMENT ON COLUMN templates.logs.event_time IS 'Время события';
COMMENT ON COLUMN templates.logs.level IS 'Уровень логирования (INFO, WARN, ERROR)';
COMMENT ON COLUMN templates.logs.message IS 'Сообщение лога';
COMMENT ON COLUMN templates.logs.metadata IS 'Дополнительные данные события';

INSERT INTO templates.users (username, email, full_name)
SELECT 
    'user_' || g,
    'user_' || g || '@example.com',
    'User ' || g
FROM generate_series(1, 1000) AS g;

INSERT INTO templates.orders (user_id, total_amount, status, shipping_address)
SELECT 
    (random() * 999 + 1)::INT,
    round((random() * 900 + 100)::NUMERIC, 2),
    (ARRAY['new', 'processing', 'done', 'cancelled'])[ceil(random() * 4)],
    'Address #' || g
FROM generate_series(1, 1000) AS g;

INSERT INTO templates.logs (level, message, metadata)
SELECT 
    (ARRAY['INFO', 'WARN', 'ERROR'])[ceil(random() * 3)],
    'Generated log #' || g,
    jsonb_build_object('event', 'test', 'index', g)
FROM generate_series(1, 1000) AS g;
