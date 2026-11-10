package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"little-petals/internal/models"
)

func TestPostRepository_Validation(t *testing.T) {
	repo := NewPostRepository(nil)

	t.Run("nil post", func(t *testing.T) {
		err := repo.Create(nil)
		if !errors.Is(err, ErrInvalidPost) {
			t.Fatalf("expected ErrInvalidPost, got %v", err)
		}
	})

	t.Run("zero or negative user id", func(t *testing.T) {
		for _, uid := range []int{0, -1, -5} {
			err := repo.Create(&models.Post{
				UserID: uid,
				Title:  "Valid Title",
			})
			if !errors.Is(err, ErrInvalidPost) {
				t.Fatalf("for uid %d, expected ErrInvalidPost, got %v", uid, err)
			}
		}
	})

	t.Run("blank title", func(t *testing.T) {
		err := repo.Create(&models.Post{
			UserID: 1,
			Title:  "   ",
		})
		if !errors.Is(err, ErrInvalidPost) {
			t.Fatalf("expected ErrInvalidPost, got %v", err)
		}
	})

	t.Run("unsupported post type", func(t *testing.T) {
		err := repo.Create(&models.Post{
			UserID:   1,
			Title:    "Valid Title",
			PostType: "video", // not blog, poem, or song
		})
		if !errors.Is(err, ErrInvalidPost) {
			t.Fatalf("expected ErrInvalidPost, got %v", err)
		}
	})

	t.Run("find by non-positive id", func(t *testing.T) {
		for _, id := range []int{0, -1, -10} {
			p, err := repo.FindByID(id)
			if !errors.Is(err, ErrPostNotFound) {
				t.Fatalf("for id %d, expected ErrPostNotFound, got %v", id, err)
			}
			if p != nil {
				t.Fatalf("for id %d, expected nil post, got %+v", id, p)
			}
		}
	})

	t.Run("list by non-positive user id returns empty slice", func(t *testing.T) {
		for _, uid := range []int{0, -1, -10} {
			posts, err := repo.ListByUserID(uid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(posts) != 0 {
				t.Fatalf("expected empty slice, got %d items", len(posts))
			}
		}
	})
}

func TestPostRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	postRepo := NewPostRepository(db)

	// Create test user for foreign key relation
	author := &models.User{
		Name:         "Post Author",
		Email:        "author@example.com",
		PasswordHash: "secret-hash",
	}
	if err := userRepo.Create(author); err != nil {
		t.Fatalf("failed to create author user: %v", err)
	}

	secondAuthor := &models.User{
		Name:         "Second Author",
		Email:        "author2@example.com",
		PasswordHash: "secret-hash-2",
	}
	if err := userRepo.Create(secondAuthor); err != nil {
		t.Fatalf("failed to create second author: %v", err)
	}

	t.Run("create post with foreign key violation", func(t *testing.T) {
		p := &models.Post{
			UserID: 999999, // non-existent user
			Title:  "Orphan Post",
		}
		err := postRepo.Create(p)
		if !errors.Is(err, ErrInvalidPost) {
			t.Fatalf("expected ErrInvalidPost for FK violation, got %v", err)
		}
	})

	t.Run("create and find post by ID (all fields and nullable fields)", func(t *testing.T) {
		post := &models.Post{
			UserID:      author.ID,
			Title:       "A Song of Morning",
			Description: "A gentle morning acoustic piece.",
			Content:     "Verse 1:\nThe sun rises on the petals...",
			PostType:    "song",
			FileURL:     "/static/uploads/songs/morning.mp3",
		}

		if err := postRepo.Create(post); err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		if post.ID <= 0 {
			t.Fatalf("expected positive post ID, got %d", post.ID)
		}
		if post.CreatedAt.IsZero() {
			t.Fatal("expected non-zero CreatedAt")
		}
		if post.UpdatedAt.IsZero() {
			t.Fatal("expected non-zero UpdatedAt")
		}
		if post.FilterType != "song" {
			t.Errorf("expected FilterType to sync with PostType, got %q", post.FilterType)
		}

		// Retrieve by ID
		found, err := postRepo.FindByID(post.ID)
		if err != nil {
			t.Fatalf("failed to find post by ID: %v", err)
		}

		if found.ID != post.ID {
			t.Errorf("ID mismatch: got %d, want %d", found.ID, post.ID)
		}
		if found.UserID != author.ID {
			t.Errorf("UserID mismatch: got %d, want %d", found.UserID, author.ID)
		}
		if found.Title != "A Song of Morning" {
			t.Errorf("Title mismatch: got %q, want %q", found.Title, "A Song of Morning")
		}
		if found.Description != "A gentle morning acoustic piece." {
			t.Errorf("Description mismatch: got %q, want %q", found.Description, "A gentle morning acoustic piece.")
		}
		if found.Content != "Verse 1:\nThe sun rises on the petals..." {
			t.Errorf("Content mismatch: got %q, want %q", found.Content, "Verse 1:\nThe sun rises on the petals...")
		}
		if found.PostType != "song" {
			t.Errorf("PostType mismatch: got %q, want %q", found.PostType, "song")
		}
		if found.FilterType != "song" {
			t.Errorf("FilterType mismatch: got %q, want %q", found.FilterType, "song")
		}
		if found.FileURL != "/static/uploads/songs/morning.mp3" {
			t.Errorf("FileURL mismatch: got %q, want %q", found.FileURL, "/static/uploads/songs/morning.mp3")
		}
	})

	t.Run("create post defaults post_type to blog and handles FilterType", func(t *testing.T) {
		post := &models.Post{
			UserID: author.ID,
			Title:  "Default Type Post",
		}
		if err := postRepo.Create(post); err != nil {
			t.Fatalf("failed to create post: %v", err)
		}
		if post.PostType != "blog" || post.FilterType != "blog" {
			t.Errorf("expected default 'blog', got post_type=%q, filter_type=%q", post.PostType, post.FilterType)
		}

		// Now create with FilterType set instead of PostType
		post2 := &models.Post{
			UserID:     author.ID,
			Title:      "Poem via FilterType",
			FilterType: "poem",
		}
		if err := postRepo.Create(post2); err != nil {
			t.Fatalf("failed to create post: %v", err)
		}
		if post2.PostType != "poem" || post2.FilterType != "poem" {
			t.Errorf("expected 'poem', got post_type=%q, filter_type=%q", post2.PostType, post2.FilterType)
		}
	})

	t.Run("find by non-existent post ID", func(t *testing.T) {
		_, err := postRepo.FindByID(999999)
		if !errors.Is(err, ErrPostNotFound) {
			t.Fatalf("expected ErrPostNotFound, got %v", err)
		}
	})

	t.Run("ListAll with filters and sorting", func(t *testing.T) {
		cleanTables(t, db)
		// Re-create authors
		_ = userRepo.Create(author)
		_ = userRepo.Create(secondAuthor)

		// Seed posts with distinct timestamps
		baseTime := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)

		blog1 := &models.Post{
			UserID:    author.ID,
			Title:     "First Blog",
			PostType:  "blog",
			CreatedAt: baseTime.Add(1 * time.Hour),
		}
		poem1 := &models.Post{
			UserID:    author.ID,
			Title:     "First Poem",
			PostType:  "poem",
			CreatedAt: baseTime.Add(2 * time.Hour),
		}
		song1 := &models.Post{
			UserID:    secondAuthor.ID,
			Title:     "First Song",
			PostType:  "song",
			CreatedAt: baseTime.Add(3 * time.Hour),
		}
		blog2 := &models.Post{
			UserID:    secondAuthor.ID,
			Title:     "Second Blog",
			PostType:  "blog",
			CreatedAt: baseTime.Add(4 * time.Hour),
		}

		for _, p := range []*models.Post{blog1, poem1, song1, blog2} {
			if err := postRepo.Create(p); err != nil {
				t.Fatalf("failed to seed post %q: %v", p.Title, err)
			}
		}

		// 1. ListAll with empty filter returns all 4 posts in DESC created_at order
		allPosts, err := postRepo.ListAll("")
		if err != nil {
			t.Fatalf("ListAll failed: %v", err)
		}
		if len(allPosts) != 4 {
			t.Fatalf("expected 4 posts, got %d", len(allPosts))
		}
		if allPosts[0].Title != "Second Blog" || allPosts[3].Title != "First Blog" {
			t.Errorf("ordering incorrect: first=%q, last=%q", allPosts[0].Title, allPosts[3].Title)
		}

		// 2. ListAll("all") returns all 4 posts
		allPostsExplicit, err := postRepo.ListAll("all")
		if err != nil {
			t.Fatalf("ListAll('all') failed: %v", err)
		}
		if len(allPostsExplicit) != 4 {
			t.Fatalf("expected 4 posts with 'all', got %d", len(allPostsExplicit))
		}

		// 3. ListAll("blog") returns 2 blogs
		blogs, err := postRepo.ListAll("blog")
		if err != nil {
			t.Fatalf("ListAll('blog') failed: %v", err)
		}
		if len(blogs) != 2 {
			t.Fatalf("expected 2 blogs, got %d", len(blogs))
		}
		for _, b := range blogs {
			if b.PostType != "blog" {
				t.Errorf("expected post_type 'blog', got %q", b.PostType)
			}
		}

		// 4. ListAll("poem") returns 1 poem
		poems, err := postRepo.ListAll("poem")
		if err != nil {
			t.Fatalf("ListAll('poem') failed: %v", err)
		}
		if len(poems) != 1 || poems[0].Title != "First Poem" {
			t.Fatalf("expected 1 poem 'First Poem', got %+v", poems)
		}

		// 5. ListAll with unknown type returns empty slice
		none, err := postRepo.ListAll("nonexistent")
		if err != nil {
			t.Fatalf("ListAll unknown failed: %v", err)
		}
		if len(none) != 0 {
			t.Fatalf("expected empty slice, got %d items", len(none))
		}
	})

	t.Run("ListByUserID", func(t *testing.T) {
		cleanTables(t, db)
		_ = userRepo.Create(author)
		_ = userRepo.Create(secondAuthor)

		p1 := &models.Post{UserID: author.ID, Title: "Author Post 1", PostType: "blog"}
		p2 := &models.Post{UserID: author.ID, Title: "Author Post 2", PostType: "poem"}
		p3 := &models.Post{UserID: secondAuthor.ID, Title: "Second Author Post", PostType: "song"}

		for _, p := range []*models.Post{p1, p2, p3} {
			if err := postRepo.Create(p); err != nil {
				t.Fatalf("failed to create post: %v", err)
			}
		}

		authorPosts, err := postRepo.ListByUserID(author.ID)
		if err != nil {
			t.Fatalf("ListByUserID failed: %v", err)
		}
		if len(authorPosts) != 2 {
			t.Fatalf("expected 2 posts for author, got %d", len(authorPosts))
		}

		secondPosts, err := postRepo.ListByUserID(secondAuthor.ID)
		if err != nil {
			t.Fatalf("ListByUserID second author failed: %v", err)
		}
		if len(secondPosts) != 1 {
			t.Fatalf("expected 1 post for second author, got %d", len(secondPosts))
		}

		// Non-existent user returns empty slice without error
		emptyPosts, err := postRepo.ListByUserID(999999)
		if err != nil {
			t.Fatalf("ListByUserID empty failed: %v", err)
		}
		if len(emptyPosts) != 0 {
			t.Fatalf("expected empty slice, got %d", len(emptyPosts))
		}
	})

	t.Run("context propagation and cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		ctxRepo := postRepo.WithContext(ctx)

		post := &models.Post{
			UserID:   author.ID,
			Title:    "Cancelled Post",
			PostType: "blog",
		}

		err := ctxRepo.Create(post)
		if err == nil {
			t.Fatal("expected error on cancelled context, got nil")
		}
		if !strings.Contains(err.Error(), "context canceled") && !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context cancellation error, got %v", err)
		}
	})
}
