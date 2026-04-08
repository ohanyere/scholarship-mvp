package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/example/scholarship-platform/services/api/internal/ai"
	"github.com/example/scholarship-platform/services/api/internal/auth"
	"github.com/example/scholarship-platform/services/api/internal/cache"
	"github.com/example/scholarship-platform/services/api/internal/config"
	"github.com/example/scholarship-platform/services/api/internal/health"
	httpapi "github.com/example/scholarship-platform/services/api/internal/http"
	"github.com/example/scholarship-platform/services/api/internal/http/handlers"
	"github.com/example/scholarship-platform/services/api/internal/metrics"
	"github.com/example/scholarship-platform/services/api/internal/repository"
	"github.com/example/scholarship-platform/services/api/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

type App struct {
	config     config.Config
	logger     *slog.Logger
	db         *sql.DB
	redis      *redis.Client
	httpServer *http.Server
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger := newLogger(cfg.LogLevel)
	authManager := auth.NewManager(cfg.JWTSecret, cfg.JWTAccessTokenTTL)
	metricsCollector := metrics.New()
	aiClient := ai.NewNoopClient()
	db, err := openPostgres(cfg)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}
	redisClient, err := openRedis(cfg)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open redis connection: %w", err)
	}
	scholarshipCache := cache.NewRedisScholarshipCache(redisClient, cfg.ScholarshipCacheTTL)

	userRepo := repository.NewPostgresUserRepository(db)
	scholarshipRepo := repository.NewPostgresScholarshipRepository(db)
	bookmarkRepo := repository.NewPostgresBookmarkRepository(db)
	applicationRepo := repository.NewPostgresApplicationRepository(db)

	services := service.NewServices(
		userRepo,
		scholarshipRepo,
		bookmarkRepo,
		applicationRepo,
		scholarshipCache,
		aiClient,
		authManager,
	)

	healthService := health.NewService(
		health.NewChecker("api", func(context.Context) error { return nil }),
		health.NewChecker("users_repository", userRepo.Ping),
		health.NewChecker("scholarships_repository", scholarshipRepo.Ping),
		health.NewChecker("bookmarks_repository", bookmarkRepo.Ping),
		health.NewChecker("applications_repository", applicationRepo.Ping),
		health.NewChecker("redis", func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		}),
	)

	handlerSet := handlers.New(healthService, services, aiClient)
	router := httpapi.NewRouter(httpapi.RouterParams{
		Logger:         logger,
		AuthManager:    authManager,
		Metrics:        metricsCollector,
		Handlers:       handlerSet,
		RequestTimeout: cfg.RequestTimeout,
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: router,
	}

	return &App{
		config:     cfg,
		logger:     logger,
		db:         db,
		redis:      redisClient,
		httpServer: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.logger.Info("starting api server", "addr", a.httpServer.Addr, "env", a.config.AppEnv)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
		defer cancel()

		a.logger.Info("shutting down api server")
		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}

		if err := a.redis.Close(); err != nil {
			return err
		}

		return a.db.Close()
	case err := <-errCh:
		return err
	}
}

func openPostgres(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.PostgresDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func openRedis(cfg config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}

func newLogger(level string) *slog.Logger {
	var slogLevel slog.Level

	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	}))
}
