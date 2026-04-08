package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/scholarship-platform/services/api/internal/ai"
	"github.com/example/scholarship-platform/services/api/internal/auth"
	"github.com/example/scholarship-platform/services/api/internal/health"
	"github.com/example/scholarship-platform/services/api/internal/http/response"
	"github.com/example/scholarship-platform/services/api/internal/repository"
	"github.com/example/scholarship-platform/services/api/internal/service"
)

type HandlerSet struct {
	health   *health.Service
	services service.Services
	aiClient ai.Client
}

func New(healthService *health.Service, services service.Services, aiClient ai.Client) HandlerSet {
	return HandlerSet{
		health:   healthService,
		services: services,
		aiClient: aiClient,
	}
}

func (h HandlerSet) RegisterPublicRoutes(r chi.Router, metricsHandler http.Handler) {
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)
	r.Handle("/metrics", metricsHandler)

	r.Route("/scholarships", func(r chi.Router) {
		r.Get("/", h.ListScholarships)
		r.Get("/{id}", h.GetScholarship)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
	})
}

func (h HandlerSet) RegisterUserRoutes(r chi.Router) {
	r.Route("/me", func(r chi.Router) {
		r.Get("/", h.Me)
		r.Route("/bookmarks", func(r chi.Router) {
			r.Get("/", h.ListBookmarks)
			r.Post("/{scholarshipId}", h.CreateBookmark)
			r.Delete("/{scholarshipId}", h.DeleteBookmark)
		})
		r.Route("/applications", func(r chi.Router) {
			r.Get("/", h.ListApplications)
			r.Post("/", h.CreateApplication)
			r.Patch("/{id}", h.UpdateApplication)
		})
	})
}

func (h HandlerSet) RegisterAdminRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Route("/scholarships", func(r chi.Router) {
			r.Post("/", h.CreateScholarship)
			r.Patch("/{id}", h.UpdateScholarship)
			r.Delete("/{id}", h.DeleteScholarship)
		})
	})
}

func (h HandlerSet) RegisterAIRoutes(r chi.Router) {
	r.Route("/ai", func(r chi.Router) {
		r.Post("/motivation-letter-draft", h.GenerateMotivationLetterDraft)
	})
}

func (h HandlerSet) Healthz(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h HandlerSet) Readyz(w http.ResponseWriter, r *http.Request) {
	checks, ready := h.health.Ready(r.Context())
	if !ready {
		response.JSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready",
			"checks": checks,
		})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"status": "ready",
		"checks": checks,
	})
}

func (h HandlerSet) ListScholarships(w http.ResponseWriter, r *http.Request) {
	scholarships, err := h.services.Scholarships.ListPublishedScholarships(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list scholarships")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"items": scholarships,
	})
}

func (h HandlerSet) GetScholarship(w http.ResponseWriter, r *http.Request) {
	scholarship, err := h.services.Scholarships.GetPublishedScholarshipByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrScholarshipNotFound):
			response.Error(w, http.StatusNotFound, "scholarship not found")
		case errors.Is(err, service.ErrValidation):
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "failed to get scholarship")
		}
		return
	}

	response.JSON(w, http.StatusOK, scholarship)
}

func (h HandlerSet) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.services.Auth.Register(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		response.Error(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h HandlerSet) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.services.Auth.Login(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			response.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, auth.ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "invalid email or password")
		default:
			response.Error(w, http.StatusInternalServerError, "failed to login")
		}
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h HandlerSet) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.services.User.GetMe(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			response.Error(w, http.StatusUnauthorized, "user not found")
		case errors.Is(err, service.ErrValidation):
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "failed to fetch current user")
		}
		return
	}

	response.JSON(w, http.StatusOK, user)
}

func (h HandlerSet) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	bookmarks, err := h.services.User.ListBookmarks(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		response.Error(w, http.StatusInternalServerError, "failed to list bookmarks")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"items": bookmarks})
}

func (h HandlerSet) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	bookmark, err := h.services.User.CreateBookmark(r.Context(), userID, chi.URLParam(r, "scholarshipId"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			response.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrScholarshipNotFound):
			response.Error(w, http.StatusNotFound, "scholarship not found")
		default:
			response.Error(w, http.StatusInternalServerError, "failed to create bookmark")
		}
		return
	}

	response.JSON(w, http.StatusCreated, bookmark)
}

func (h HandlerSet) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := h.services.User.DeleteBookmark(r.Context(), userID, chi.URLParam(r, "scholarshipId")); err != nil {
		if errors.Is(err, service.ErrValidation) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		response.Error(w, http.StatusInternalServerError, "failed to delete bookmark")
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}

func (h HandlerSet) ListApplications(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	applications, err := h.services.User.ListApplications(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		response.Error(w, http.StatusInternalServerError, "failed to list applications")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"items": applications})
}

func (h HandlerSet) CreateApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var input service.CreateApplicationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	application, err := h.services.User.CreateApplication(r.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			response.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrScholarshipNotFound):
			response.Error(w, http.StatusNotFound, "scholarship not found")
		case errors.Is(err, repository.ErrApplicationAlreadyExists):
			response.Error(w, http.StatusBadRequest, "application already exists for this scholarship")
		default:
			response.Error(w, http.StatusInternalServerError, "failed to create application")
		}
		return
	}

	response.JSON(w, http.StatusCreated, application)
}

func (h HandlerSet) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var input service.UpdateApplicationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	application, err := h.services.User.UpdateApplication(r.Context(), userID, chi.URLParam(r, "id"), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			response.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrApplicationNotFound):
			response.Error(w, http.StatusNotFound, "application not found")
		default:
			response.Error(w, http.StatusInternalServerError, "failed to update application")
		}
		return
	}

	response.JSON(w, http.StatusOK, application)
}

func (h HandlerSet) CreateScholarship(w http.ResponseWriter, r *http.Request) {
	var input service.CreateScholarshipInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	scholarship, err := h.services.Admin.CreateScholarship(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		response.Error(w, http.StatusInternalServerError, "failed to create scholarship")
		return
	}

	response.JSON(w, http.StatusCreated, scholarship)
}

func (h HandlerSet) UpdateScholarship(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusNotImplemented, map[string]string{
		"message": "update scholarship scaffold",
	})
}

func (h HandlerSet) DeleteScholarship(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusNotImplemented, map[string]string{
		"message": "delete scholarship scaffold",
	})
}

func (h HandlerSet) GenerateMotivationLetterDraft(w http.ResponseWriter, r *http.Request) {
	draft, err := h.aiClient.GenerateMotivationLetterDraft(r.Context(), ai.DraftRequest{})
	if err != nil {
		response.Error(w, http.StatusBadGateway, "ai provider unavailable")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "motivation letter draft scaffold",
		"draft":   draft,
	})
}

func currentUserID(r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		return "", false
	}

	return claims.UserID, true
}
