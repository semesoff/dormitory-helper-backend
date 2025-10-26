-- +goose Up
-- +goose StatementBegin

-- Создание таблицы пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL
);

-- Время жизни пользователя
CREATE TABLE users_time_live (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    time_live TIMESTAMP NOT NULL DEFAULT NOW() + interval '7 days' -- Время жизни
);

-- Создание типа ролей
CREATE TYPE user_role AS ENUM ('user', 'admin');

-- Создание таблицы ролей
CREATE TABLE roles (
    id SERIAL REFERENCES users(id) ON DELETE CASCADE,
    role user_role NOT NULL
);

-- Создание таблицы записи на прачечную с start_time и end_time
CREATE TABLE laundry_bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    CONSTRAINT ch_start_end_times CHECK (end_time > start_time),
    CONSTRAINT ch_time_interval CHECK (end_time - start_time <= interval '2 hours')
);

-- Создание таблицы кухни
CREATE TABLE kitchen_bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    CONSTRAINT ch_kitchen_start_end_times CHECK (end_time > start_time),
    CONSTRAINT ch_kitchen_time_interval CHECK (end_time - start_time <= interval '3 hour')
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS kitchen_bookings;
DROP TABLE IF EXISTS laundry_bookings;
DROP TABLE IF EXISTS roles;
DROP TYPE IF EXISTS user_role;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
