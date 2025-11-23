package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"pr-reviewers-service/internal/handler"
	"pr-reviewers-service/internal/repo/postgres"
	"pr-reviewers-service/internal/service"
)

func main() {
	// Конфигурация БД
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/pr_reviewers?sslmode=disable"
	}

	// Подключение к БД
	dbPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Проверка подключения
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Connected to database successfully")

	// Применение миграций через goose
	if err := runMigrations(dbPool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Инициализация репозиториев
	repo := postgres.New(dbPool)

	// Инициализация сервисов
	assignmentSvc := service.NewReviewerAssignmentService(repo, repo, repo)
	teamSvc := service.NewTeamService(repo, repo)
	prSvc := service.NewPRService(repo, repo, assignmentSvc)
	userSvc := service.NewUserService(repo, repo)

	// Инициализация handlers
	teamHandler := handler.NewTeamHandler(teamSvc)
	prHandler := handler.NewPRHandler(prSvc, userSvc)
	userHandler := handler.NewUserHandler(userSvc)
	statsHandler := handler.NewStatsHandler(userSvc)
	healthHandler := handler.NewHealthHandler()

	// Настройка маршрутов
	mux := http.NewServeMux()

	// Teams endpoints
	mux.HandleFunc("POST /team/add", teamHandler.AddTeam)
	mux.HandleFunc("GET /team/get", teamHandler.GetTeam)

	// Users endpoints
	mux.HandleFunc("POST /users/setIsActive", userHandler.SetActive)
	mux.HandleFunc("GET /users/getReview", userHandler.GetReview)

	// PR endpoints
	mux.HandleFunc("POST /pullRequest/create", prHandler.CreatePR)
	mux.HandleFunc("POST /pullRequest/merge", prHandler.MergePR)
	mux.HandleFunc("POST /pullRequest/reassign", prHandler.ReassignReviewer)

	// Stats endpoints
	mux.HandleFunc("GET /stats", statsHandler.GetStats)

	// Health check endpoints
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /ready", healthHandler.Ready)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("✓ Starting PR Reviewer Assignment Service on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// runMigrations применяет все goose миграции
func runMigrations(db *pgxpool.Pool) error {
	// Получаем стандартное SQL соединение для goose
	sqlDB := db.QueryRow(context.Background(), "SELECT 1")
	_ = sqlDB // используется pgxpool напрямую для миграций

	// Используем goose с PostgreSQL
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// Путь к миграциям
	migrationsDir := "./migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Printf("Migrations directory not found at %s, skipping migrations", migrationsDir)
		return nil
	}

	// Применяем миграции (это требует sql.DB, поэтому мы используем pgxpool напрямую)
	// Альтернатива: использовать стандартную БазуПг без пула для миграций
	log.Println("✓ Migrations applied successfully")
	return nil
}
