package database

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestChecksum(t *testing.T) {
	c1 := checksum([]byte("hello world"))
	c2 := checksum([]byte("hello world"))
	if c1 != c2 {
		t.Errorf("expected deterministic checksum, got %q != %q", c1, c2)
	}

	empty := checksum(nil)
	if empty != emptyChecksum {
		t.Errorf("emptyChecksum mismatch: got %q, want %q", empty, emptyChecksum)
	}
}

func TestMigrateValidation(t *testing.T) {
	ctx := context.Background()

	if _, err := Migrate(ctx, nil, fstest.MapFS{}); err == nil {
		t.Error("expected error when db is nil, got nil")
	}

	if _, err := Migrate(ctx, nil, nil); err == nil {
		t.Error("expected error when db and files are nil, got nil")
	}
}
