package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"little-petals/internal/models"
)

var (
	// ErrUserNotFound is returned when the requested user is not present in the database.
	ErrUserNotFound = errors.New("repository: user not found")

	// ErrDuplicateEmail is returned when attempting to insert a user whose email is already registered.
	ErrDuplicateEmail = errors.New("repository: email already registered")

	// ErrInvalidUser is returned when a user struct is nil or missing required values.
	ErrInvalidUser = errors.New("repository: invalid user payload")
)

// UserRepository defines the contract for persisting and retrieving user entities.
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id int) (*models.User, error)
	WithContext(ctx context.Context) UserRepository
}

// PostgresUserRepository provides a PostgreSQL-backed implementation of UserRepository.
type PostgresUserRepository struct {
	db  *sql.DB
	ctx context.Context
}

// Ensure PostgresUserRepository satisfies UserRepository at compile time.
var _ UserRepository = (*PostgresUserRepository)(nil)

// NewUserRepository constructs a new PostgresUserRepository.
func NewUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:  db,
		ctx: context.Background(),
	}
}

// WithContext returns a shallow copy of the repository bound to the supplied context.
func (r *PostgresUserRepository) WithContext(ctx context.Context) UserRepository {
	if ctx == nil {
		ctx = context.Background()
	}
	return &PostgresUserRepository{
		db:  r.db,
		ctx: ctx,
	}
}

func (r *PostgresUserRepository) currentContext() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

// Create inserts a new user record into PostgreSQL using parameterized placeholders.
// If user.Name is empty, it derives a fallback name from the email address.
func (r *PostgresUserRepository) Create(user *models.User) error {
	return r.CreateWithContext(r.currentContext(), user)
}

// CreateWithContext inserts a new user using the specified context.
func (r *PostgresUserRepository) CreateWithContext(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("%w: user is nil", ErrInvalidUser)
	}

	email := strings.TrimSpace(user.Email)
	if email == "" {
		return fmt.Errorf("%w: email cannot be blank", ErrInvalidUser)
	}
	user.Email = email

	passwordHash := strings.TrimSpace(user.PasswordHash)
	if passwordHash == "" {
		return fmt.Errorf("%w: password_hash cannot be blank", ErrInvalidUser)
	}
	user.PasswordHash = passwordHash

	name := strings.TrimSpace(user.Name)
	if name == "" {
		parts := strings.Split(email, "@")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			name = strings.TrimSpace(parts[0])
		} else {
			name = "user"
		}
	}
	user.Name = name

	createdAt := user.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	query := `
		INSERT INTO users (name, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query, user.Name, user.Email, user.PasswordHash, createdAt).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: %s", ErrDuplicateEmail, user.Email)
		}
		return fmt.Errorf("user_repo: create: %w", err)
	}

	user.UpdatedAt = user.CreatedAt
	return nil
}

// FindByEmail retrieves a user by their email address using case-insensitive matching.
func (r *PostgresUserRepository) FindByEmail(email string) (*models.User, error) {
	return r.FindByEmailWithContext(r.currentContext(), email)
}

// FindByEmailWithContext retrieves a user by their email address using the specified context.
func (r *PostgresUserRepository) FindByEmailWithContext(ctx context.Context, email string) (*models.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, ErrUserNotFound
	}

	query := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user_repo: find by email: %w", err)
	}

	user.UpdatedAt = user.CreatedAt
	return &user, nil
}

// FindByID retrieves a user by their unique primary key ID.
func (r *PostgresUserRepository) FindByID(id int) (*models.User, error) {
	return r.FindByIDWithContext(r.currentContext(), id)
}

// FindByIDWithContext retrieves a user by their unique primary key ID using the specified context.
func (r *PostgresUserRepository) FindByIDWithContext(ctx context.Context, id int) (*models.User, error) {
	if id <= 0 {
		return nil, ErrUserNotFound
	}

	query := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user_repo: find by id: %w", err)
	}

	user.UpdatedAt = user.CreatedAt
	return &user, nil
}
