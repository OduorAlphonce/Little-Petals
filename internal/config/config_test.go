package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var configEnvKeys = []string{
	"ENV_FILE",
	"PORT",
	"DB_CONN_STR",
	"DATABASE_URL",
	"STATIC_DIR",
	"UPLOAD_DIR",
	"SONGS_UPLOAD_DIR",
	"COVERS_UPLOAD_DIR",
	"BOOKS_UPLOAD_DIR",
	"DB_MAX_OPEN_CONNS",
	"DB_MAX_IDLE_CONNS",
	"DB_CONN_MAX_LIFETIME",
	"DB_CONN_MAX_IDLE_TIME",
	"DB_PING_TIMEOUT",
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range configEnvKeys {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
}

func TestLoadRequiresDatabaseConnectionString(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())

	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded without DB_CONN_STR")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())
	t.Setenv("DB_CONN_STR", "postgres://localhost/littlepetals")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != defaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, defaultPort)
	}
	if cfg.StaticDir != defaultStaticDir {
		t.Errorf("StaticDir = %q, want %q", cfg.StaticDir, defaultStaticDir)
	}
	if cfg.UploadsDir != defaultUploadsDir {
		t.Errorf("UploadsDir = %q, want %q", cfg.UploadsDir, defaultUploadsDir)
	}
	if cfg.SongsUploadDir != filepath.Join(defaultUploadsDir, "songs") {
		t.Errorf("SongsUploadDir = %q", cfg.SongsUploadDir)
	}
	if cfg.CoversUploadDir != filepath.Join(defaultUploadsDir, "covers") {
		t.Errorf("CoversUploadDir = %q", cfg.CoversUploadDir)
	}
	if cfg.BooksUploadDir != filepath.Join(defaultUploadsDir, "books") {
		t.Errorf("BooksUploadDir = %q", cfg.BooksUploadDir)
	}
	if cfg.DBMaxOpenConns != defaultMaxOpen {
		t.Errorf("DBMaxOpenConns = %d, want %d", cfg.DBMaxOpenConns, defaultMaxOpen)
	}
	if cfg.DBMaxIdleConns != defaultMaxIdle {
		t.Errorf("DBMaxIdleConns = %d, want %d", cfg.DBMaxIdleConns, defaultMaxIdle)
	}
	if cfg.DBConnMaxLifetime != defaultConnLife {
		t.Errorf("DBConnMaxLifetime = %v, want %v", cfg.DBConnMaxLifetime, defaultConnLife)
	}
	if cfg.DBConnMaxIdleTime != defaultConnIdle {
		t.Errorf("DBConnMaxIdleTime = %v, want %v", cfg.DBConnMaxIdleTime, defaultConnIdle)
	}
	if cfg.DBPingTimeout != defaultPingTimeout {
		t.Errorf("DBPingTimeout = %v, want %v", cfg.DBPingTimeout, defaultPingTimeout)
	}
}

func TestLoadDotEnvAndEnvironmentPrecedence(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	t.Chdir(dir)

	envFile := []byte(
		"PORT=9090\n" +
			"DB_CONN_STR=postgres://from-dotenv/littlepetals\n" +
			"STATIC_DIR=public\n" +
			"UPLOAD_DIR=public/files\n" +
			"DB_MAX_OPEN_CONNS=12\n" +
			"DB_MAX_IDLE_CONNS=4\n" +
			"DB_CONN_MAX_LIFETIME=15m\n" +
			"DB_CONN_MAX_IDLE_TIME=3m\n" +
			"DB_PING_TIMEOUT=2s\n",
	)
	if err := os.WriteFile(filepath.Join(dir, ".env"), envFile, 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Setenv("PORT", "8081")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 8081 {
		t.Errorf("Port = %d, want environment override 8081", cfg.Port)
	}
	if cfg.DBConnStr != "postgres://from-dotenv/littlepetals" {
		t.Errorf("DBConnStr = %q", cfg.DBConnStr)
	}
	if cfg.SongsUploadDir != filepath.Join("public/files", "songs") {
		t.Errorf("SongsUploadDir = %q", cfg.SongsUploadDir)
	}
	if cfg.DBMaxOpenConns != 12 {
		t.Errorf("DBMaxOpenConns = %d, want 12", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 4 {
		t.Errorf("DBMaxIdleConns = %d, want 4", cfg.DBMaxIdleConns)
	}
	if cfg.DBConnMaxLifetime != 15*time.Minute {
		t.Errorf("DBConnMaxLifetime = %v", cfg.DBConnMaxLifetime)
	}
	if cfg.DBConnMaxIdleTime != 3*time.Minute {
		t.Errorf("DBConnMaxIdleTime = %v", cfg.DBConnMaxIdleTime)
	}
	if cfg.DBPingTimeout != 2*time.Second {
		t.Errorf("DBPingTimeout = %v", cfg.DBPingTimeout)
	}
}

func TestLoadFallsBackToDatabaseURL(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())
	t.Setenv("DATABASE_URL", "postgres://localhost/fallback")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBConnStr != "postgres://localhost/fallback" {
		t.Errorf("DBConnStr = %q", cfg.DBConnStr)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "non-numeric port", key: "PORT", value: "http"},
		{name: "port out of range", key: "PORT", value: "70000"},
		{name: "zero max open connections", key: "DB_MAX_OPEN_CONNS", value: "0"},
		{name: "idle connections exceed open connections", key: "DB_MAX_IDLE_CONNS", value: "100"},
		{name: "invalid duration", key: "DB_CONN_MAX_LIFETIME", value: "soon"},
		{name: "negative duration", key: "DB_CONN_MAX_IDLE_TIME", value: "-1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv(t)
			t.Chdir(t.TempDir())
			t.Setenv("DB_CONN_STR", "postgres://localhost/littlepetals")
			t.Setenv(tt.key, tt.value)

			if _, err := Load(); err == nil {
				t.Fatalf("Load() succeeded with %s=%q", tt.key, tt.value)
			}
		})
	}
}

func TestLoadExplicitEnvFileMustExist(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())
	t.Setenv("ENV_FILE", "missing.env")
	t.Setenv("DB_CONN_STR", "postgres://localhost/littlepetals")

	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded with a missing explicit ENV_FILE")
	}
}

func TestUploadDirsDeduplicates(t *testing.T) {
	cfg := &Config{
		UploadsDir:      "uploads",
		SongsUploadDir:  "uploads/songs",
		CoversUploadDir: "uploads/songs",
		BooksUploadDir:  "",
	}

	got := cfg.UploadDirs()
	want := []string{"uploads", "uploads/songs"}
	if len(got) != len(want) {
		t.Fatalf("UploadDirs() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("UploadDirs()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
