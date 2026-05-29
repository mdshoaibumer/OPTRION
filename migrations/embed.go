package migrations

import "embed"

// FS contains all SQL migration files embedded at build time.
//
//go:embed *.sql
var FS embed.FS
