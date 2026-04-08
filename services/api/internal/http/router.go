package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/example/scholarship-platform/services/api/internal/auth"
	"github.com/example/scholarship-platform/services/api/internal/http/handlers"
	appmiddleware "github.com/example/scholarship-platform/services/api/internal/http/middleware"
	"github.com/example/scholarship-platform/services/api/internal/metrics"
)

type RouterParams struct {
	Logger         *slog.Logger
	AuthManager    *auth.Manager
	Metrics        *metrics.Metrics
	Handlers       handlers.HandlerSet
	RequestTimeout time.Duration
}

func NewRouter(params RouterParams) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(params.RequestTimeout))
	router.Use(appmiddleware.RequestLogger(params.Logger))
	router.Use(params.Metrics.Middleware)

	params.Handlers.RegisterPublicRoutes(router, params.Metrics.Handler())

	router.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth(params.AuthManager))
		params.Handlers.RegisterUserRoutes(r)
		params.Handlers.RegisterAIRoutes(r)
	})

	router.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth(params.AuthManager))
		r.Use(appmiddleware.RequireRole(auth.RoleAdmin))
		params.Handlers.RegisterAdminRoutes(r)
	})

	return router
}
