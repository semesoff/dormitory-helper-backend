CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('student', 'admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bookings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL CHECK (resource_type IN ('laundry', 'kitchen')),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (start_at < end_at)
);

CREATE INDEX IF NOT EXISTS idx_bookings_resource_time ON bookings(resource_type, start_at, end_at);

CREATE TABLE IF NOT EXISTS repairs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_repairs_user_status ON repairs(user_id, status);

CREATE TABLE IF NOT EXISTS exchange_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    category TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('give', 'exchange', 'sell')),
    description TEXT NOT NULL,
    contact TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exchange_items_created_at ON exchange_items(created_at DESC);

CREATE TABLE IF NOT EXISTS announcements (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    priority TEXT NOT NULL CHECK (priority IN ('high', 'medium', 'low')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_announcements_created_at ON announcements(created_at DESC);

CREATE TABLE IF NOT EXISTS polls (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    end_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_polls_status_end_at ON polls(status, end_at);

CREATE TABLE IF NOT EXISTS poll_options (
    id BIGSERIAL PRIMARY KEY,
    poll_id BIGINT NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    option_text TEXT NOT NULL,
    votes BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS poll_votes (
    id BIGSERIAL PRIMARY KEY,
    poll_id BIGINT NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    option_id BIGINT NOT NULL REFERENCES poll_options(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (poll_id, user_id)
);

CREATE TABLE IF NOT EXISTS faq_entries (
    id BIGSERIAL PRIMARY KEY,
    topic TEXT NOT NULL,
    keywords TEXT[] NOT NULL DEFAULT '{}',
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_faq_entries_topic ON faq_entries(topic);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('booking', 'repair', 'poll')),
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    dedupe_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_read_created ON notifications(user_id, is_read, created_at DESC);

CREATE TABLE IF NOT EXISTS notification_settings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    bookings_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    repairs_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    polls_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO users (username, password_hash, role)
VALUES
    ('admin', '773d8f92f07dec1e40c76ff36e0079d4bb41063ef7c447c9f6d0b5ef778fd840', 'admin'),
    ('student', 'd5ba78e10ea9604a296ca2082f0eec186ebadf73a6f87b6d85617dba34df4790', 'student')
ON CONFLICT (username) DO NOTHING;

INSERT INTO notification_settings (user_id)
SELECT id
FROM users
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO faq_entries (topic, keywords, question, answer)
VALUES
    ('rules', ARRAY['тишина','шум','время'], 'Какие часы тишины действуют в общежитии?', 'Часы тишины действуют с 23:00 до 07:00 в будни и с 23:00 до 09:00 в выходные.'),
    ('laundry', ARRAY['прачечная','стирка','машинка'], 'Как пользоваться прачечной?', 'Прачечная бронируется через приложение. Максимальная длительность брони 2 часа.'),
    ('repairs', ARRAY['ремонт','заявка','мастер'], 'Как оставить заявку на ремонт?', 'Создайте заявку с указанием комнаты, категории и описания проблемы.')
ON CONFLICT DO NOTHING;
