-- migrations/00001_init_schema.sql
-- +goose Up
-- +goose StatementBegin

-- Таблица команд
CREATE TABLE IF NOT EXISTS teams (
    team_name   VARCHAR(255) PRIMARY KEY,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
    );

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    user_id     VARCHAR(255) PRIMARY KEY,
    username    VARCHAR(255) NOT NULL,
    team_name   VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
    );

-- Таблица Pull Request'ов
CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id   VARCHAR(255) PRIMARY KEY,
    pull_request_name TEXT         NOT NULL,
    author_id         VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status            VARCHAR(50)  NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    created_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    merged_at         TIMESTAMP    NULL
    );

-- Связка PR ↔ Ревьюверы (многие-ко-многим)
CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id     VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    assigned_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, reviewer_id)
    );

-- === Индексы для производительности ===
CREATE INDEX IF NOT EXISTS idx_users_team_active ON users(team_name, is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_users_team        ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_prs_author        ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_prs_status        ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_prs_created_desc  ON pull_requests(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr   ON pr_reviewers(pull_request_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_pr_reviewers_reviewer;
DROP INDEX IF EXISTS idx_pr_reviewers_pr;
DROP INDEX IF EXISTS idx_prs_created_desc;
DROP INDEX IF EXISTS idx_prs_status;
DROP INDEX IF EXISTS idx_prs_author;
DROP INDEX IF EXISTS idx_users_team;
DROP INDEX IF EXISTS idx_users_team_active;

DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

-- +goose StatementEnd