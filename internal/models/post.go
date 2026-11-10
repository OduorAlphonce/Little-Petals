package models

import "time"

// Post represents a creative post (blog, poem, or song) stored in the PostgreSQL database.
type Post struct {
	ID          int       `db:"id" json:"id"`
	UserID      int       `db:"user_id" json:"user_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Content     string    `db:"content" json:"content"`
	PostType    string    `db:"post_type" json:"post_type"`
	FilterType  string    `db:"filter_type" json:"filter_type"`
	FileURL     string    `db:"file_url" json:"file_url,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
