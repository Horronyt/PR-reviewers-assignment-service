// tests/integration_test.go
package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://127.0.0.1:8080"

type IntegrationTest struct {
	db     *pgxpool.Pool
	client *http.Client
	sqlDB  *sql.DB
}

// Отключаем параллельный запуск всех интеграционных тестов
func TestMain(m *testing.M) {
	// Запускаем сервер ДО выполнения тестов
	startTestServerOnce()

	// Даём возможность завершить сервер по особым сигналам
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		serverMu.Lock()
		if serverCmd != nil && serverCmd.Process != nil {
			fmt.Println("\nReceived interrupt, killing test server...")
			serverCmd.Process.Kill()
		}
		serverMu.Unlock()
		os.Exit(0)
	}()

	// Запускаем тесты
	code := m.Run()

	// После завершения тестов — грациозно убиваем сервер
	serverMu.Lock()
	if serverCmd != nil && serverCmd.Process != nil {
		fmt.Println("Tests finished, killing test server...")
		serverCmd.Process.Kill()
		serverCmd.Wait() // важно! освобождаем ресурсы
	}
	serverMu.Unlock()

	os.Exit(code)
}

var serverCmd *exec.Cmd
var serverOnce sync.Once
var serverMu sync.Mutex

func startTestServerOnce() {
	serverOnce.Do(func() {
		fmt.Println("Starting test server...")

		testDSN := "postgres://user:password@localhost:5433/pr_reviewers_test?sslmode=disable"
		os.Setenv("DATABASE_URL", testDSN)
		os.Setenv("PORT", "8080")

		cmd := exec.Command("go", "run", "./cmd/api/...")
		cmd.Dir = findProjectRootOrDie(nil) // теперь можно без t
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(),
			"DATABASE_URL="+testDSN,
			"PORT=8080",
		)

		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start test server: %v\n", err)
			os.Exit(1)
		}

		serverMu.Lock()
		serverCmd = cmd
		serverMu.Unlock()

		// Ждём здоровья
		if !waitForHealth(30 * time.Second) {
			fmt.Fprintln(os.Stderr, "Server did not become healthy in time")
			cmd.Process.Kill()
			os.Exit(1)
		}

		fmt.Println("Test server is ready!")
	})
}

func findProjectRootOrDie(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found — are you in the project?")
		}
		dir = parent
	}
}

func waitForHealth(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://127.0.0.1:8080/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

func New(t *testing.T) *IntegrationTest {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	tester := &IntegrationTest{
		db:     pool,
		client: &http.Client{Timeout: 10 * time.Second},
	}

	t.Cleanup(func() {
		tester.truncateAllTables(t)
		pool.Close()
	})

	tester.truncateAllTables(t)

	return tester
}

func (it *IntegrationTest) truncateAllTables(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := it.db.Exec(ctx, `
        TRUNCATE TABLE pr_reviewers, pull_requests, teams, users RESTART IDENTITY CASCADE
    `)
	if err != nil {
		t.Logf("TRUNCATE warning: %v", err)
	}
}

// Оставляем Close только для пула — но он теперь в t.Cleanup
func (it *IntegrationTest) Close() {
}

// Post и Get — без изменений
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
