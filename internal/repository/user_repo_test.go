package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"little-petals/internal/models"
)

func TestUserRepository_Validation(t *testing.T) {
	repo := NewUserRepository(nil)

	t.Run("nil user payload", func(t *testing.T) {
		err := repo.Create(nil)
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
	})

	t.Run("blank email", func(t *testing.T) {
		err := repo.Create(&models.User{
			Email:        "   ",
			PasswordHash: "secret123",
		})
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
	})

	t.Run("blank password hash", func(t *testing.T) {
		err := repo.Create(&models.User{
			Email:        "user@example.com",
			PasswordHash: "   ",
		})
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
	})

	t.Run("find by blank email", func(t *testing.T) {
		user, err := repo.FindByEmail("   ")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
		if user != nil {
			t.Fatalf("expected nil user, got %+v", user)
		}
	})

	t.Run("find by non-positive id", func(t *testing.T) {
		for _, id := range []int{0, -1, -42} {
			user, err := repo.FindByID(id)
			if !errors.Is(err, ErrUserNotFound) {
				t.Fatalf("for id %d, expected ErrUserNotFound, got %v", id, err)
			}
			if user != nil {
				t.Fatalf("for id %d, expected nil user, got %+v", id, user)
			}
		}
	})
}

func TestUserRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	t.Run("create and retrieve user with custom name", func(t *testing.T) {
		user := &models.User{
			Name:         "Jane Doe",
			Email:        "jane.doe@example.com",
			PasswordHash: "hashed-pwd-jane",
		}

		if err := repo.Create(user); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		if user.ID <= 0 {
			t.Fatalf("expected positive user ID, got %d", user.ID)
		}
		if user.CreatedAt.IsZero() {
			t.Fatal("expected non-zero CreatedAt")
		}
		if user.UpdatedAt.IsZero() {
			t.Fatal("expected non-zero UpdatedAt")
		}

		// Retrieval by ID
		foundByID, err := repo.FindByID(user.ID)
		if err != nil {
			t.Fatalf("failed to find user by ID: %v", err)
		}
		if foundByID.ID != user.ID {
			t.Errorf("ID mismatch: got %d, want %d", foundByID.ID, user.ID)
		}
		if foundByID.Name != "Jane Doe" {
			t.Errorf("Name mismatch: got %q, want %q", foundByID.Name, "Jane Doe")
		}
		if foundByID.Email != "jane.doe@example.com" {
			t.Errorf("Email mismatch: got %q, want %q", foundByID.Email, "jane.doe@example.com")
		}
		if foundByID.PasswordHash != "hashed-pwd-jane" {
			t.Errorf("PasswordHash mismatch: got %q, want %q", foundByID.PasswordHash, "hashed-pwd-jane")
		}

		// Retrieval by Email (case-insensitive)
		foundByEmail, err := repo.FindByEmail("JANE.DOE@EXAMPLE.COM")
		if err != nil {
			t.Fatalf("failed to find user by case-insensitive email: %v", err)
		}
		if foundByEmail.ID != user.ID {
			t.Errorf("Email search ID mismatch: got %d, want %d", foundByEmail.ID, user.ID)
		}
	})

	t.Run("create user with empty name derives fallback from email", func(t *testing.T) {
		user := &models.User{
			Email:        "author_creative@example.com",
			PasswordHash: "hashed-author-pwd",
		}

		if err := repo.Create(user); err != nil {
			t.Fatalf("failed to create user with empty name: %v", err)
		}

		if user.Name != "author_creative" {
			t.Errorf("expected fallback name %q, got %q", "author_creative", user.Name)
		}

		found, err := repo.FindByID(user.ID)
		if err != nil {
			t.Fatalf("failed to find user: %v", err)
		}
		if found.Name != "author_creative" {
			t.Errorf("persisted name mismatch: got %q, want %q", found.Name, "author_creative")
		}
	})

	t.Run("create duplicate email returns ErrDuplicateEmail", func(t *testing.T) {
		u1 := &models.User{
			Name:         "User One",
			Email:        "unique@example.com",
			PasswordHash: "pass-1",
		}
		if err := repo.Create(u1); err != nil {
			t.Fatalf("failed to create initial user: %v", err)
		}

		u2 := &models.User{
			Name:         "User Two",
			Email:        "UNIQUE@example.com", // case-insensitive clash
			PasswordHash: "pass-2",
		}
		err := repo.Create(u2)
		if !errors.Is(err, ErrDuplicateEmail) {
			t.Fatalf("expected ErrDuplicateEmail, got %v", err)
		}
	})

	t.Run("find by non-existent email or ID", func(t *testing.T) {
		_, err := repo.FindByEmail("nonexistent@example.com")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound for email, got %v", err)
		}

		_, err = repo.FindByID(999999)
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound for ID, got %v", err)
		}
	})

	t.Run("context propagation and cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		ctxRepo := repo.WithContext(ctx)

		user := &models.User{
			Name:         "Cancelled User",
			Email:        "cancelled@example.com",
			PasswordHash: "hash-pwd",
		}

		err := ctxRepo.Create(user)
		if err == nil {
			t.Fatal("expected error on cancelled context, got nil")
		}
		if !strings.Contains(err.Error(), "context canceled") && !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context cancellation error, got %v", err)
		}
	})
}
