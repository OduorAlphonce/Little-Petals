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
	// ErrPostNotFound is returned when the requested post cannot be found.
	ErrPostNotFound = errors.New("repository: post not found")

	// ErrInvalidPost is returned when a post struct fails validation.
	ErrInvalidPost = errors.New("repository: invalid post payload")
)

// PostRepository defines the contract for persisting and querying post entities.
type PostRepository interface {
	Create(post *models.Post) error
	FindByID(id int) (*models.Post, error)
	ListAll(filterType string) ([]*models.Post, error)
	ListByUserID(userID int) ([]*models.Post, error)
	WithContext(ctx context.Context) PostRepository
}

// PostgresPostRepository provides a PostgreSQL-backed implementation of PostRepository.
type PostgresPostRepository struct {
	db  *sql.DB
	ctx context.Context
}

// Ensure PostgresPostRepository satisfies PostRepository at compile time.
var _ PostRepository = (*PostgresPostRepository)(nil)

// NewPostRepository constructs a new PostgresPostRepository.
func NewPostRepository(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{
		db:  db,
		ctx: context.Background(),
	}
}

// WithContext returns a shallow copy of the repository bound to the supplied context.
func (r *PostgresPostRepository) WithContext(ctx context.Context) PostRepository {
	if ctx == nil {
		ctx = context.Background()
	}
	return &PostgresPostRepository{
		db:  r.db,
		ctx: ctx,
	}
}

func (r *PostgresPostRepository) currentContext() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

// Create inserts a new post record into PostgreSQL using parameterized placeholders.
func (r *PostgresPostRepository) Create(post *models.Post) error {
	return r.CreateWithContext(r.currentContext(), post)
}

// CreateWithContext inserts a new post using the specified context.
func (r *PostgresPostRepository) CreateWithContext(ctx context.Context, post *models.Post) error {
	if post == nil {
		return fmt.Errorf("%w: post is nil", ErrInvalidPost)
	}

	if post.UserID <= 0 {
		return fmt.Errorf("%w: user_id must be greater than 0", ErrInvalidPost)
	}

	title := strings.TrimSpace(post.Title)
	if title == "" {
		return fmt.Errorf("%w: title cannot be blank", ErrInvalidPost)
	}
	post.Title = title

	postType := strings.ToLower(strings.TrimSpace(post.PostType))
	if postType == "" {
		postType = strings.ToLower(strings.TrimSpace(post.FilterType))
	}
	if postType == "" {
		postType = "blog"
	}

	switch postType {
	case "blog", "poem", "song":
		// Allowed post types
	default:
		return fmt.Errorf("%w: unsupported post type %q", ErrInvalidPost, postType)
	}
	post.PostType = postType
	post.FilterType = postType

	createdAt := post.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	query := `
		INSERT INTO posts (user_id, title, description, content, post_type, file_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		post.UserID,
		post.Title,
		post.Description,
		post.Content,
		post.PostType,
		post.FileURL,
		createdAt,
	).Scan(&post.ID, &post.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return fmt.Errorf("%w: user_id %d does not exist", ErrInvalidPost, post.UserID)
		}
		return fmt.Errorf("post_repo: create: %w", err)
	}

	post.UpdatedAt = post.CreatedAt
	return nil
}

// FindByID retrieves a single post by its ID.
func (r *PostgresPostRepository) FindByID(id int) (*models.Post, error) {
	return r.FindByIDWithContext(r.currentContext(), id)
}

// FindByIDWithContext retrieves a single post by its ID using the specified context.
func (r *PostgresPostRepository) FindByIDWithContext(ctx context.Context, id int) (*models.Post, error) {
	if id <= 0 {
		return nil, ErrPostNotFound
	}

	query := `
		SELECT id, user_id, title, description, content, post_type, file_url, created_at
		FROM posts
		WHERE id = $1
	`

	var p models.Post
	var desc, content, fileURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.UserID,
		&p.Title,
		&desc,
		&content,
		&p.PostType,
		&fileURL,
		&p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("post_repo: find by id: %w", err)
	}

	p.Description = desc.String
	p.Content = content.String
	p.FileURL = fileURL.String
	p.FilterType = p.PostType
	p.UpdatedAt = p.CreatedAt

	return &p, nil
}

// ListAll retrieves posts optionally filtered by post/filter type, ordered by newest first.
// If filterType is empty or "all", all posts are returned.
func (r *PostgresPostRepository) ListAll(filterType string) ([]*models.Post, error) {
	return r.ListAllWithContext(r.currentContext(), filterType)
}

// ListAllWithContext retrieves posts optionally filtered by post type using the specified context.
func (r *PostgresPostRepository) ListAllWithContext(ctx context.Context, filterType string) ([]*models.Post, error) {
	filter := strings.ToLower(strings.TrimSpace(filterType))
	var query string
	var args []any

	if filter == "" || filter == "all" {
		query = `
			SELECT id, user_id, title, description, content, post_type, file_url, created_at
			FROM posts
			ORDER BY created_at DESC
		`
	} else {
		query = `
			SELECT id, user_id, title, description, content, post_type, file_url, created_at
			FROM posts
			WHERE LOWER(post_type) = LOWER($1)
			ORDER BY created_at DESC
		`
		args = append(args, filter)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("post_repo: list all: %w", err)
	}
	defer rows.Close()

	posts := make([]*models.Post, 0)
	for rows.Next() {
		var p models.Post
		var desc, content, fileURL sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&desc,
			&content,
			&p.PostType,
			&fileURL,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("post_repo: scan post: %w", err)
		}
		p.Description = desc.String
		p.Content = content.String
		p.FileURL = fileURL.String
		p.FilterType = p.PostType
		p.UpdatedAt = p.CreatedAt
		posts = append(posts, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("post_repo: iterate posts: %w", err)
	}

	return posts, nil
}

// ListByUserID retrieves all posts authored by a specific user, ordered by newest first.
func (r *PostgresPostRepository) ListByUserID(userID int) ([]*models.Post, error) {
	return r.ListByUserIDWithContext(r.currentContext(), userID)
}

// ListByUserIDWithContext retrieves all posts authored by a specific user using the specified context.
func (r *PostgresPostRepository) ListByUserIDWithContext(ctx context.Context, userID int) ([]*models.Post, error) {
	if userID <= 0 {
		return []*models.Post{}, nil
	}

	query := `
		SELECT id, user_id, title, description, content, post_type, file_url, created_at
		FROM posts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("post_repo: list by user id: %w", err)
	}
	defer rows.Close()

	posts := make([]*models.Post, 0)
	for rows.Next() {
		var p models.Post
		var desc, content, fileURL sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&desc,
			&content,
			&p.PostType,
			&fileURL,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("post_repo: scan post: %w", err)
		}
		p.Description = desc.String
		p.Content = content.String
		p.FileURL = fileURL.String
		p.FilterType = p.PostType
		p.UpdatedAt = p.CreatedAt
		posts = append(posts, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("post_repo: iterate posts: %w", err)
	}

	return posts, nil
}
