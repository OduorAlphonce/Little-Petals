// Package migrations contains the embedded SQL migrations applied at startup.
package migrations

import "embed"

// Files holds every .sql migration in this directory.
//
//go:embed *.sql
var Files embed.FS
