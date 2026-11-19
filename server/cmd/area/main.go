// Command area is the single AREA server binary. It runs in one of three
// roles, chosen with the ROLE environment variable (or the -role flag):
//
//   - api:       REST API, OAuth/OIDC callbacks and webhooks
//   - scheduler: single-leader polling loop
//   - worker:    workflow run executor
//
// See docs/ARCHITECTURE.md (modular monolith, one binary, three roles).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/auth"
	"github.com/polcortadellas/AREA/server/internal/db"
	"github.com/polcortadellas/AREA/server/internal/mail"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	role := flag.String("role", envOr("ROLE", "api"), "runtime role: api, scheduler or worker")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("starting", "role", *role)

	switch *role {
	case "api":
		return runAPI(ctx, envOr("API_ADDR", ":8080"))
	case "scheduler", "worker":
		// Filled in by M-11 / M-12; for now the process just stays up.
		<-ctx.Done()
		return nil
	default:
		return fmt.Errorf("unknown role %q (want api, scheduler or worker)", *role)
	}
}

func runAPI(ctx context.Context, addr string) error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return errors.New("DATABASE_URL is not set")
	}
	if err := db.Migrate(dbURL); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer pool.Close()

	jwtTTL, err := time.ParseDuration(envOr("JWT_TTL", "1h"))
	if err != nil {
		return fmt.Errorf("JWT_TTL: %w", err)
	}
	tokens, err := auth.NewTokens(os.Getenv("JWT_SECRET"), jwtTTL)
	if err != nil {
		return err
	}
	sender := mail.SMTP{
		Addr: net.JoinHostPort(envOr("SMTP_HOST", "localhost"), envOr("SMTP_PORT", "1025")),
		From: envOr("SMTP_FROM", "AREA <no-reply@area.local>"),
	}
	authSvc, err := auth.NewService(pool, tokens, sender, envOr("WEB_BASE_URL", "http://localhost:8081"))
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	authSvc.Routes(mux)
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("pong\n"))
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("api listening", "addr", addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
