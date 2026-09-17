package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AofArthit21/Hospital-information-system/internal/auth"
	"github.com/AofArthit21/Hospital-information-system/internal/config"
	"github.com/AofArthit21/Hospital-information-system/internal/his"
	"github.com/AofArthit21/Hospital-information-system/internal/httpapi"
	postgresrepo "github.com/AofArthit21/Hospital-information-system/internal/repository/postgres"
	"github.com/AofArthit21/Hospital-information-system/internal/service"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--healthcheck" {
		client := http.Client{Timeout: 2 * time.Second}
		response, err := client.Get("http://127.0.0.1:8080/healthz")
		if err != nil || response.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		_ = response.Body.Close()
		return
	}
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		return err
	}

	hospitals := postgresrepo.NewHospitalRepository(db)
	staff := postgresrepo.NewStaffRepository(db)
	patients := postgresrepo.NewPatientRepository(db)
	tokens := auth.NewManager(cfg.JWTSecret, cfg.JWTTTL)
	authService := service.NewAuthService(hospitals, staff, tokens)
	patientService := service.NewPatientService(hospitals, patients, his.NewClient(cfg.HISTimeout))
	router := httpapi.NewRouter(httpapi.NewHandler(authService, patientService), tokens)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	serveErr := make(chan error, 1)
	go func() {
		slog.Info("API listening", "address", cfg.HTTPAddr)
		serveErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
