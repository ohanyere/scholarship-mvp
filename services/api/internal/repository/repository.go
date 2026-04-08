package repository

import (
	"context"
	"errors"
	"time"
)

var ErrScholarshipNotFound = errors.New("scholarship not found")
var ErrUserNotFound = errors.New("user not found")
var ErrUserEmailTaken = errors.New("user email already exists")
var ErrApplicationNotFound = errors.New("application not found")
var ErrApplicationAlreadyExists = errors.New("application already exists")

type UserRepository interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, params CreateUserParams) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
	FullName     string
	Role         string
}

type Scholarship struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Provider       string    `json:"provider"`
	Description    string    `json:"description"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	DeadlineAt     time.Time `json:"deadline_at"`
	Eligibility    string    `json:"eligibility"`
	ApplicationURL string    `json:"application_url"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateScholarshipParams struct {
	Title          string
	Provider       string
	Description    string
	Amount         int64
	Currency       string
	DeadlineAt     time.Time
	Eligibility    string
	ApplicationURL string
	Status         string
}

type ScholarshipRepository interface {
	Ping(ctx context.Context) error
	CreateScholarship(ctx context.Context, params CreateScholarshipParams) (Scholarship, error)
	ListPublishedScholarships(ctx context.Context) ([]Scholarship, error)
	GetPublishedScholarshipByID(ctx context.Context, id string) (Scholarship, error)
}

type Bookmark struct {
	UserID        string    `json:"user_id"`
	ScholarshipID string    `json:"scholarship_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type BookmarkRepository interface {
	Ping(ctx context.Context) error
	CreateBookmark(ctx context.Context, userID, scholarshipID string) (Bookmark, error)
	ListBookmarks(ctx context.Context, userID string) ([]Bookmark, error)
	DeleteBookmark(ctx context.Context, userID, scholarshipID string) error
}

type Application struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ScholarshipID string    `json:"scholarship_id"`
	Status        string    `json:"status"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateApplicationParams struct {
	UserID        string
	ScholarshipID string
	Status        string
	Notes         string
}

type UpdateApplicationParams struct {
	ID     string
	UserID string
	Status *string
	Notes  *string
}

type ApplicationRepository interface {
	Ping(ctx context.Context) error
	CreateApplication(ctx context.Context, params CreateApplicationParams) (Application, error)
	ListApplications(ctx context.Context, userID string) ([]Application, error)
	UpdateApplication(ctx context.Context, params UpdateApplicationParams) (Application, error)
}

type NoopUserRepository struct{}
type NoopBookmarkRepository struct{}
type NoopApplicationRepository struct{}

func NewNoopUserRepository() *NoopUserRepository {
	return &NoopUserRepository{}
}

func NewNoopBookmarkRepository() *NoopBookmarkRepository {
	return &NoopBookmarkRepository{}
}

func NewNoopApplicationRepository() *NoopApplicationRepository {
	return &NoopApplicationRepository{}
}

func (r *NoopUserRepository) Ping(context.Context) error {
	return nil
}

func (r *NoopUserRepository) CreateUser(context.Context, CreateUserParams) (User, error) {
	return User{}, nil
}

func (r *NoopUserRepository) GetUserByEmail(context.Context, string) (User, error) {
	return User{}, ErrUserNotFound
}

func (r *NoopUserRepository) GetUserByID(context.Context, string) (User, error) {
	return User{}, ErrUserNotFound
}

func (r *NoopBookmarkRepository) Ping(context.Context) error {
	return nil
}

func (r *NoopBookmarkRepository) CreateBookmark(context.Context, string, string) (Bookmark, error) {
	return Bookmark{}, nil
}

func (r *NoopBookmarkRepository) ListBookmarks(context.Context, string) ([]Bookmark, error) {
	return nil, nil
}

func (r *NoopBookmarkRepository) DeleteBookmark(context.Context, string, string) error {
	return nil
}

func (r *NoopApplicationRepository) Ping(context.Context) error {
	return nil
}

func (r *NoopApplicationRepository) CreateApplication(context.Context, CreateApplicationParams) (Application, error) {
	return Application{}, nil
}

func (r *NoopApplicationRepository) ListApplications(context.Context, string) ([]Application, error) {
	return nil, nil
}

func (r *NoopApplicationRepository) UpdateApplication(context.Context, UpdateApplicationParams) (Application, error) {
	return Application{}, ErrApplicationNotFound
}
