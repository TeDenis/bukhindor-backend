-- +goose Up
-- Создание таблицы для сброса паролей
CREATE TABLE IF NOT EXISTS password_resets (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Создание индексов для password_resets
CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at);

-- Обновление таблицы users для соответствия новой структуре
-- Удаляем старые колонки если они существуют
PRAGMA foreign_keys=off;

-- Создаем временную таблицу с новой структурой
CREATE TABLE IF NOT EXISTS users_new (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Копируем данные из старой таблицы (если есть)
INSERT OR IGNORE INTO users_new (id, email, name, password_hash, is_active, created_at, updated_at)
SELECT 
    id,
    email,
    COALESCE(first_name || ' ' || last_name, username) as name,
    password_hash,
    is_active,
    created_at,
    updated_at
FROM users;

-- Удаляем старую таблицу и переименовываем новую
DROP TABLE IF EXISTS users;
ALTER TABLE users_new RENAME TO users;

PRAGMA foreign_keys=on;

-- Создание индексов для обновленной таблицы users
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- +goose Down
-- Удаление таблицы password_resets
DROP TABLE IF EXISTS password_resets;

-- Восстановление старой структуры таблицы users
PRAGMA foreign_keys=off;

CREATE TABLE IF NOT EXISTS users_old (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    role TEXT DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Копируем данные обратно
INSERT OR IGNORE INTO users_old (id, email, username, password_hash, first_name, last_name, is_active, role, created_at, updated_at)
SELECT 
    id,
    email,
    name as username,
    password_hash,
    name as first_name,
    '' as last_name,
    is_active,
    'user' as role,
    created_at,
    updated_at
FROM users;

DROP TABLE IF EXISTS users;
ALTER TABLE users_old RENAME TO users;

PRAGMA foreign_keys=on;

-- Восстановление индексов
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at); 