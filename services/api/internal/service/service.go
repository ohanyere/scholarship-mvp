package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/example/scholarship-platform/services/api/internal/ai"
	"github.com/example/scholarship-platform/services/api/internal/auth"
	"github.com/example/scholarship-platform/services/api/internal/cache"
	"github.com/example/scholarship-platform/services/api/internal/repository"
)

var ErrValidation = errors.New("validation error")

type CreateScholarshipInput struct {
	Title          string    `json:"title"`
	Provider       string    `json:"provider"`
	Description    string    `json:"description"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	DeadlineAt     time.Time `json:"deadline_at"`
	Eligibility    string    `json:"eligibility"`
	ApplicationURL string    `json:"application_url"`
	Status         string    `json:"status"`
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResult struct {
	AccessToken string          `json:"access_token"`
	User        repository.User `json:"user"`
}

type CreateApplicationInput struct {
	ScholarshipID string `json:"scholarship_id"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}

type UpdateApplicationInput struct {
	Status *string `json:"status"`
	Notes  *string `json:"notes"`
}

type ScholarshipService struct {
	scholarships repository.ScholarshipRepository
	cache        cache.ScholarshipCache
}

type AuthService struct {
	users repository.UserRepository
	auth  *auth.Manager
}

type UserService struct {
	users        repository.UserRepository
	bookmarks    repository.BookmarkRepository
	applications repository.ApplicationRepository
}

type AdminService struct {
	scholarships repository.ScholarshipRepository
	cache        cache.ScholarshipCache
}

type AIService struct {
	client ai.Client
}

type Services struct {
	Scholarships *ScholarshipService
	Auth         *AuthService
	User         *UserService
	Admin        *AdminService
	AI           *AIService
}

func NewServices(
	userRepo repository.UserRepository,
	scholarshipRepo repository.ScholarshipRepository,
	bookmarkRepo repository.BookmarkRepository,
	applicationRepo repository.ApplicationRepository,
	scholarshipCache cache.ScholarshipCache,
	aiClient ai.Client,
	authManager *auth.Manager,
) Services {
	return Services{
		Scholarships: &ScholarshipService{
			scholarships: scholarshipRepo,
			cache:        scholarshipCache,
		},
		Auth: &AuthService{
			users: userRepo,
			auth:  authManager,
		},
		User: &UserService{
			users:        userRepo,
			bookmarks:    bookmarkRepo,
			applications: applicationRepo,
		},
		Admin: &AdminService{
			scholarships: scholarshipRepo,
			cache:        scholarshipCache,
		},
		AI: &AIService{
			client: aiClient,
		},
	}
}

func (s *ScholarshipService) ListPublishedScholarships(ctx context.Context) ([]repository.Scholarship, error) {
	cached, hit, err := s.cache.GetScholarshipList(ctx)
	if err == nil && hit {
		var scholarships []repository.Scholarship
		if unmarshalErr := json.Unmarshal(cached, &scholarships); unmarshalErr == nil {
			return scholarships, nil
		}
	}

	scholarships, err := s.scholarships.ListPublishedScholarships(ctx)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(scholarships)
	if err == nil {
		_ = s.cache.SetScholarshipList(ctx, payload)
	}

	return scholarships, nil
}

func (s *ScholarshipService) GetPublishedScholarshipByID(ctx context.Context, id string) (repository.Scholarship, error) {
	if strings.TrimSpace(id) == "" {
		return repository.Scholarship{}, fmt.Errorf("%w: scholarship id is required", ErrValidation)
	}

	cached, hit, err := s.cache.GetScholarshipDetail(ctx, id)
	if err == nil && hit {
		var scholarship repository.Scholarship
		if unmarshalErr := json.Unmarshal(cached, &scholarship); unmarshalErr == nil {
			return scholarship, nil
		}
	}

	scholarship, err := s.scholarships.GetPublishedScholarshipByID(ctx, id)
	if err != nil {
		return repository.Scholarship{}, err
	}

	payload, err := json.Marshal(scholarship)
	if err == nil {
		_ = s.cache.SetScholarshipDetail(ctx, id, payload)
	}

	return scholarship, nil
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)
	fullName := strings.TrimSpace(input.FullName)

	switch {
	case email == "":
		return AuthResult{}, fmt.Errorf("%w: email is required", ErrValidation)
	case !isValidEmail(email):
		return AuthResult{}, fmt.Errorf("%w: email must be valid", ErrValidation)
	case fullName == "":
		return AuthResult{}, fmt.Errorf("%w: full_name is required", ErrValidation)
	case len(password) < 8:
		return AuthResult{}, fmt.Errorf("%w: password must be at least 8 characters", ErrValidation)
	}

	passwordHash, err := s.auth.HashPassword(password)
	if err != nil {
		return AuthResult{}, err
	}

	user, err := s.users.CreateUser(ctx, repository.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         string(auth.RoleUser),
	})
	if err != nil {
		if errors.Is(err, repository.ErrUserEmailTaken) {
			return AuthResult{}, fmt.Errorf("%w: email is already registered", ErrValidation)
		}

		return AuthResult{}, err
	}

	token, err := s.auth.GenerateToken(auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   auth.Role(user.Role),
	})
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		AccessToken: token,
		User:        sanitizeUser(user),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)

	switch {
	case email == "":
		return AuthResult{}, fmt.Errorf("%w: email is required", ErrValidation)
	case password == "":
		return AuthResult{}, fmt.Errorf("%w: password is required", ErrValidation)
	}

	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AuthResult{}, auth.ErrInvalidCredentials
		}

		return AuthResult{}, err
	}

	if err := s.auth.VerifyPassword(password, user.PasswordHash); err != nil {
		return AuthResult{}, err
	}

	token, err := s.auth.GenerateToken(auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   auth.Role(user.Role),
	})
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		AccessToken: token,
		User:        sanitizeUser(user),
	}, nil
}

func (s *UserService) GetMe(ctx context.Context, userID string) (repository.User, error) {
	if strings.TrimSpace(userID) == "" {
		return repository.User{}, fmt.Errorf("%w: user id is required", ErrValidation)
	}

	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return repository.User{}, err
	}

	return sanitizeUser(user), nil
}

func (s *UserService) CreateBookmark(ctx context.Context, userID, scholarshipID string) (repository.Bookmark, error) {
	if strings.TrimSpace(userID) == "" {
		return repository.Bookmark{}, fmt.Errorf("%w: user id is required", ErrValidation)
	}

	if strings.TrimSpace(scholarshipID) == "" {
		return repository.Bookmark{}, fmt.Errorf("%w: scholarship id is required", ErrValidation)
	}

	return s.bookmarks.CreateBookmark(ctx, userID, scholarshipID)
}

func (s *UserService) ListBookmarks(ctx context.Context, userID string) ([]repository.Bookmark, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrValidation)
	}

	return s.bookmarks.ListBookmarks(ctx, userID)
}

func (s *UserService) DeleteBookmark(ctx context.Context, userID, scholarshipID string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("%w: user id is required", ErrValidation)
	}

	if strings.TrimSpace(scholarshipID) == "" {
		return fmt.Errorf("%w: scholarship id is required", ErrValidation)
	}

	return s.bookmarks.DeleteBookmark(ctx, userID, scholarshipID)
}

func (s *UserService) CreateApplication(ctx context.Context, userID string, input CreateApplicationInput) (repository.Application, error) {
	if strings.TrimSpace(userID) == "" {
		return repository.Application{}, fmt.Errorf("%w: user id is required", ErrValidation)
	}

	scholarshipID := strings.TrimSpace(input.ScholarshipID)
	status := strings.TrimSpace(input.Status)
	notes := strings.TrimSpace(input.Notes)

	if scholarshipID == "" {
		return repository.Application{}, fmt.Errorf("%w: scholarship_id is required", ErrValidation)
	}

	if status == "" {
		status = "draft"
	}

	if status != "draft" && status != "submitted" && status != "withdrawn" {
		return repository.Application{}, fmt.Errorf("%w: status must be draft, submitted, or withdrawn", ErrValidation)
	}

	return s.applications.CreateApplication(ctx, repository.CreateApplicationParams{
		UserID:        userID,
		ScholarshipID: scholarshipID,
		Status:        status,
		Notes:         notes,
	})
}

func (s *UserService) ListApplications(ctx context.Context, userID string) ([]repository.Application, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrValidation)
	}

	return s.applications.ListApplications(ctx, userID)
}

func (s *UserService) UpdateApplication(ctx context.Context, userID, applicationID string, input UpdateApplicationInput) (repository.Application, error) {
	if strings.TrimSpace(userID) == "" {
		return repository.Application{}, fmt.Errorf("%w: user id is required", ErrValidation)
	}

	if strings.TrimSpace(applicationID) == "" {
		return repository.Application{}, fmt.Errorf("%w: application id is required", ErrValidation)
	}

	if input.Status == nil && input.Notes == nil {
		return repository.Application{}, fmt.Errorf("%w: at least one field must be provided", ErrValidation)
	}

	var status *string
	if input.Status != nil {
		normalizedStatus := strings.TrimSpace(*input.Status)
		if normalizedStatus != "draft" && normalizedStatus != "submitted" && normalizedStatus != "withdrawn" {
			return repository.Application{}, fmt.Errorf("%w: status must be draft, submitted, or withdrawn", ErrValidation)
		}
		status = &normalizedStatus
	}

	var notes *string
	if input.Notes != nil {
		normalizedNotes := strings.TrimSpace(*input.Notes)
		notes = &normalizedNotes
	}

	return s.applications.UpdateApplication(ctx, repository.UpdateApplicationParams{
		ID:     applicationID,
		UserID: userID,
		Status: status,
		Notes:  notes,
	})
}

func (s *AdminService) CreateScholarship(ctx context.Context, input CreateScholarshipInput) (repository.Scholarship, error) {
	title := strings.TrimSpace(input.Title)
	provider := strings.TrimSpace(input.Provider)
	description := strings.TrimSpace(input.Description)
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	eligibility := strings.TrimSpace(input.Eligibility)
	applicationURL := strings.TrimSpace(input.ApplicationURL)
	status := strings.TrimSpace(input.Status)

	switch {
	case title == "":
		return repository.Scholarship{}, fmt.Errorf("%w: title is required", ErrValidation)
	case provider == "":
		return repository.Scholarship{}, fmt.Errorf("%w: provider is required", ErrValidation)
	case description == "":
		return repository.Scholarship{}, fmt.Errorf("%w: description is required", ErrValidation)
	case input.Amount < 0:
		return repository.Scholarship{}, fmt.Errorf("%w: amount must be zero or greater", ErrValidation)
	case currency == "":
		return repository.Scholarship{}, fmt.Errorf("%w: currency is required", ErrValidation)
	case input.DeadlineAt.IsZero():
		return repository.Scholarship{}, fmt.Errorf("%w: deadline_at is required", ErrValidation)
	case eligibility == "":
		return repository.Scholarship{}, fmt.Errorf("%w: eligibility is required", ErrValidation)
	case applicationURL == "":
		return repository.Scholarship{}, fmt.Errorf("%w: application_url is required", ErrValidation)
	}

	if status == "" {
		status = "published"
	}

	if status != "draft" && status != "published" && status != "archived" {
		return repository.Scholarship{}, fmt.Errorf("%w: status must be draft, published, or archived", ErrValidation)
	}

	scholarship, err := s.scholarships.CreateScholarship(ctx, repository.CreateScholarshipParams{
		Title:          title,
		Provider:       provider,
		Description:    description,
		Amount:         input.Amount,
		Currency:       currency,
		DeadlineAt:     input.DeadlineAt,
		Eligibility:    eligibility,
		ApplicationURL: applicationURL,
		Status:         status,
	})
	if err != nil {
		return repository.Scholarship{}, err
	}

	_ = s.cache.InvalidateScholarshipList(ctx)
	_ = s.cache.InvalidateScholarship(ctx, scholarship.ID)

	return scholarship, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func sanitizeUser(user repository.User) repository.User {
	user.PasswordHash = ""
	return user
}
