CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE IF NOT EXISTS pull_request (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id),
    status pr_status NOT NULL DEFAULT 'OPEN',
    reviewer1_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewer2_id UUID REFERENCES users(id) ON DELETE SET NULL,
    need_more_reviewers BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    merged_at TIMESTAMP WITH TIME ZONE NULL
)