// cmd/server/main.go
package main

import (
	"context"
	"database/sql"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/handler"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/repo/postgres"
	"github.com/Horronyt/PR-reviewers-assignment-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/pr_reviewers?sslmode=disable"
	}

	dbPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		return err
	}
	log.Println("Connected to database")

	if err := runMigrations(dbPool); err != nil {
		return err
	}

	repo := postgres.New(dbPool)
	assignmentSvc := service.NewReviewerAssignmentService(repo, repo, repo)
	teamSvc := service.NewTeamService(repo, repo)
	prSvc := service.NewPRService(repo, repo, assignmentSvc)
	userSvc := service.NewUserService(repo, repo)

	teamHandler := handler.NewTeamHandler(teamSvc)
	prHandler := handler.NewPRHandler(prSvc, userSvc)
	userHandler := handler.NewUserHandler(userSvc, prSvc)
	statsHandler := handler.NewStatsHandler(userSvc)
	healthHandler := handler.NewHealthHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /team/add", teamHandler.AddTeam)
	mux.HandleFunc("GET /team/get", teamHandler.GetTeam)
	mux.HandleFunc("POST /users/setIsActive", userHandler.SetActive)
	mux.HandleFunc("GET /users/getReview", userHandler.GetReview)
	mux.HandleFunc("POST /pullRequest/create", prHandler.CreatePR)
	mux.HandleFunc("POST /pullRequest/merge", prHandler.MergePR)
	mux.HandleFunc("POST /pullRequest/reassign", prHandler.ReassignReviewer)
	mux.HandleFunc("GET /stats", statsHandler.GetStats)
	mux.HandleFunc("GET /stats/reviewers", statsHandler.GetReviewerStats)
	mux.HandleFunc("GET /stats/prs", statsHandler.GetPRStats)
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /ready", healthHandler.Ready)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server starting on :%s", port)

	// Graceful shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	return server.ListenAndServe()
}

func runMigrations(db *pgxpool.Pool) error {
	if err := goose.SetDialect("pgx"); err != nil {
		return err
	}

	sqlDB, err := sql.Open("pgx", db.Config().ConnString())
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return goose.Up(sqlDB, "./migrations")
}

func main() {
	if err := Run(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
