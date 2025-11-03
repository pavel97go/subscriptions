CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name TEXT        NOT NULL,
    price        INTEGER     NOT NULL CHECK (price >= 0),
    user_id      UUID        NOT NULL,
    start_month  DATE        NOT NULL,
    end_month    DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_period_order CHECK (end_month IS NULL OR end_month >= start_month)
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_set_updated_at ON subscriptions;

CREATE TRIGGER trg_set_updated_at
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_subs_user             ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subs_service          ON subscriptions(service_name);
CREATE INDEX IF NOT EXISTS idx_subs_period           ON subscriptions(start_month, end_month);
CREATE INDEX IF NOT EXISTS idx_subs_user_service     ON subscriptions(user_id, service_name);