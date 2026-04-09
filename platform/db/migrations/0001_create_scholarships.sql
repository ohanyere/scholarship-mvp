CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS scholarships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    provider TEXT NOT NULL,
    description TEXT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount >= 0),
    currency TEXT NOT NULL,
    deadline_at TIMESTAMPTZ NOT NULL,
    eligibility TEXT NOT NULL,
    application_url TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scholarships_status_deadline
    ON scholarships (status, deadline_at);
