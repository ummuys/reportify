CREATE TABLE IF NOT EXISTS identity.roles (
    role_id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

INSERT INTO identity.roles (name) VALUES 
('user'),
('admin');

COMMENT ON TABLE identity.roles IS 'Таблица с ролями';
COMMENT ON COLUMN identity.roles.role_id IS 'ID роли';
COMMENT ON COLUMN identity.roles.name IS 'Имя роли';