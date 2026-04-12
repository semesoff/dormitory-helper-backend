ALTER TABLE users
ADD COLUMN IF NOT EXISTS device_id TEXT UNIQUE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_device_id
ON users(device_id)
WHERE device_id IS NOT NULL;
