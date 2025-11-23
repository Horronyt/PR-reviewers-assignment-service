// tests/integration_test.go
package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://127.0.0.1:8080"

type IntegrationTest struct {
	db     *pgxpool.Pool
	client *http.Client
}

func New(t *testing.T) *IntegrationTest {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5433/pr_reviewers_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "failed to connect to test database")

	sqlDB, err := sql.Open("pgx", db.Config().ConnString())
	require.NoError(t, err)
	defer sqlDB.Close()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "migrations")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find migrations directory")
		}
		dir = parent
	}

	migrationsPath := filepath.Join(dir, "migrations")
	t.Logf("Applying migrations from: %s", migrationsPath)

	require.NoError(t, goose.SetDialect("pgx"))
	require.NoError(t, goose.Up(sqlDB, migrationsPath))
	t.Log("Test database ready with schema")

	return &IntegrationTest{
		db:     db,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (it *IntegrationTest) Close() {
	if it.db != nil {
		it.db.Close()
	}
}

func (it *IntegrationTest) Cleanup(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tables := []string{"pr_reviewers", "pull_requests", "users", "teams"}
	for _, table := range tables {
		_, err := it.db.Exec(ctx, "TRUNCATE TABLE "+table+" CASCADE")
		require.NoError(t, err, "failed to truncate table %s", table)
	}
}

func (it *IntegrationTest) Post(t *testing.T, endpoint string, body any) *http.Response {
	t.Helper()
	data, _ := json.Marshal(body)
	resp, err := it.client.Post(baseURL+endpoint, "application/json", bytes.NewBuffer(data))
	require.NoError(t, err, "POST %s failed", endpoint)
	return resp
}

func (it *IntegrationTest) Get(t *testing.T, endpoint string) *http.Response {
	t.Helper()
	resp, err := http.Get(baseURL + endpoint)
	require.NoError(t, err, "GET %s failed", endpoint)
	return resp
}
