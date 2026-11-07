package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort        = 8080
	defaultStaticDir   = "web/static"
	defaultUploadsDir  = "web/static/uploads"
	defaultMaxOpen     = 25
	defaultMaxIdle     = 5
	defaultConnLife    = 30 * time.Minute
	defaultConnIdle    = 5 * time.Minute
	defaultPingTimeout = 5 * time.Second
)

// Config holds every runtime setting required by the application.
// Values are read from the process environment, optionally hydrated from a
// local .env file. Environment variables always take precedence over .env.
type Config struct {
	Port      int
	DBConnStr string

	StaticDir       string
	UploadsDir      string
	SongsUploadDir  string
	CoversUploadDir string
	BooksUploadDir  string

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration
	DBPingTimeout     time.Duration
}

// Load reads the application configuration from a local .env file (if present)
// and the process environment. It returns a descriptive error when a required
// value is missing or malformed.
func Load() (*Config, error) {
	if err := loadDotEnv(); err != nil {
		return nil, err
	}

	port, err := intEnv("PORT", defaultPort, 1, 65535)
	if err != nil {
		return nil, err
	}

	dbConnStr := strings.TrimSpace(os.Getenv("DB_CONN_STR"))
	if dbConnStr == "" {
		dbConnStr = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dbConnStr == "" {
		return nil, errors.New("config: DB_CONN_STR is required")
	}

	staticDir := cleanDir(envOrDefault("STATIC_DIR", defaultStaticDir))
	uploadsDir := cleanDir(envOrDefault("UPLOAD_DIR", defaultUploadsDir))

	maxOpen, err := intEnv("DB_MAX_OPEN_CONNS", defaultMaxOpen, 1, 1000)
	if err != nil {
		return nil, err
	}
	maxIdle, err := intEnv("DB_MAX_IDLE_CONNS", defaultMaxIdle, 0, maxOpen)
	if err != nil {
		return nil, err
	}

	connLife, err := durationEnv("DB_CONN_MAX_LIFETIME", defaultConnLife)
	if err != nil {
		return nil, err
	}
	connIdle, err := durationEnv("DB_CONN_MAX_IDLE_TIME", defaultConnIdle)
	if err != nil {
		return nil, err
	}
	pingTimeout, err := durationEnv("DB_PING_TIMEOUT", defaultPingTimeout)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:      port,
		DBConnStr: dbConnStr,

		StaticDir:       staticDir,
		UploadsDir:      uploadsDir,
		SongsUploadDir:  cleanDir(envOrDefault("SONGS_UPLOAD_DIR", filepath.Join(uploadsDir, "songs"))),
		CoversUploadDir: cleanDir(envOrDefault("COVERS_UPLOAD_DIR", filepath.Join(uploadsDir, "covers"))),
		BooksUploadDir:  cleanDir(envOrDefault("BOOKS_UPLOAD_DIR", filepath.Join(uploadsDir, "books"))),

		DBMaxOpenConns:    maxOpen,
		DBMaxIdleConns:    maxIdle,
		DBConnMaxLifetime: connLife,
		DBConnMaxIdleTime: connIdle,
		DBPingTimeout:     pingTimeout,
	}, nil
}

// UploadDirs returns every directory that must exist before uploads are
// accepted. Duplicates are removed so callers can safely create each path.
func (c *Config) UploadDirs() []string {
	seen := make(map[string]struct{}, 4)
	dirs := make([]string, 0, 4)
	for _, dir := range []string{c.UploadsDir, c.SongsUploadDir, c.CoversUploadDir, c.BooksUploadDir} {
		if dir == "" {
			continue
		}
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		dirs = append(dirs, dir)
	}
	return dirs
}

func loadDotEnv() error {
	path := strings.TrimSpace(os.Getenv("ENV_FILE"))
	explicit := path != ""
	if !explicit {
		path = ".env"
	}

	err := godotenv.Load(path)
	if err == nil {
		return nil
	}
	if explicit {
		return fmt.Errorf("config: load %s: %w", path, err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config: load %s: %w", path, err)
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func cleanDir(path string) string {
	return filepath.Clean(path)
}

func intEnv(key string, fallback, min, max int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be an integer: %w", key, err)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("config: %s must be between %d and %d", key, min, max)
	}
	return value, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a valid duration (for example 5m): %w", key, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("config: %s must be greater than zero", key)
	}
	return value, nil
}
