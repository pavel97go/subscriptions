CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name TEXT        NOT NULL,
    price        INTEGER     NOT NULL,   
    user_id      UUID        NOT NULL,
    start_month  DATE        NOT NULL,      
    end_month    DATE,                      
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subs_user     ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subs_service  ON subscriptions(service_name);
CREATE INDEX IF NOT EXISTS idx_subs_period   ON subscriptions(start_month, end_month);
