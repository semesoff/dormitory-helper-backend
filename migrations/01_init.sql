-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL
);

-- Время жизни пользователя
CREATE TABLE IF NOT EXISTS users_time_live (
    user_id INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    time_live TIMESTAMP NOT NULL DEFAULT NOW() + interval '7 days'
);

-- Создание типа ролей
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('user', 'admin');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Создание таблицы ролей
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL REFERENCES users (id) ON DELETE CASCADE,
    role user_role NOT NULL
);

-- Создание таблицы записи на прачечную
CREATE TABLE IF NOT EXISTS laundry_bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users (id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    CONSTRAINT ch_start_end_times CHECK (end_time > start_time),
    CONSTRAINT ch_time_interval CHECK (
        end_time - start_time <= interval '2 hours'
    )
);

-- Создание таблицы кухни
CREATE TABLE IF NOT EXISTS kitchen_bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users (id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    CONSTRAINT ch_kitchen_start_end_times CHECK (end_time > start_time),
    CONSTRAINT ch_kitchen_time_interval CHECK (
        end_time - start_time <= interval '3 hour'
    )
);