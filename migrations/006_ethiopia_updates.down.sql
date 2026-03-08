ALTER TABLE bookings DROP COLUMN IF EXISTS payment_method;
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
DROP INDEX IF EXISTS idx_users_email_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_unique;
-- Note: Postgres does not support removing ENUM values — categories stay
