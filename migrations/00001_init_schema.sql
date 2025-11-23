-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS teams (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    team_id VARCHAR(255) NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS pull_requests (
    id VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP NULL
    );

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pr_id VARCHAR(255) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    reviewer_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pr_id, reviewer_id)
    );

-- Индексы для оптимизации запросов
CREATE INDEX idx_users_team_active ON users(team_id, is_active);
CREATE INDEX idx_users_team ON users(team_id);

CREATE INDEX idx_prs_author ON pull_requests(author_id);
CREATE INDEX idx_prs_status ON pull_requests(status);
CREATE INDEX idx_prs_created ON pull_requests(created_at DESC);

CREATE INDEX idx_pr_reviewers_pr ON pr_reviewers(pr_id);
CREATE INDEX idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_pr_reviewers_reviewer;
DROP INDEX IF EXISTS idx_pr_reviewers_pr;

DROP INDEX IF EXISTS idx_prs_created;
DROP INDEX IF EXISTS idx_prs_status;
DROP INDEX IF EXISTS idx_prs_author;

DROP INDEX IF EXISTS idx_users_team;
DROP INDEX IF EXISTS idx_users_team_active;

DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

-- +goose StatementEnd