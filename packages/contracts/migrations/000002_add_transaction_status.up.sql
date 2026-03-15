CREATE TYPE transaction_status AS ENUM ('completed', 'pending_revision');
ALTER TABLE transactions ADD COLUMN status transaction_status NOT NULL DEFAULT 'completed';
