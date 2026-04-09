CREATE TABLE IF NOT EXISTS bookmarks (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scholarship_id UUID NOT NULL REFERENCES scholarships(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, scholarship_id)
);

CREATE INDEX IF NOT EXISTS idx_bookmarks_user_id
    ON bookmarks (user_id);

CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scholarship_id UUID NOT NULL REFERENCES scholarships(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('draft', 'submitted', 'withdrawn')),
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, scholarship_id)
);

CREATE INDEX IF NOT EXISTS idx_applications_user_id
    ON applications (user_id);
