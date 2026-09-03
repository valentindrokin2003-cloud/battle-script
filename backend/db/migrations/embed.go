// Package migrations embeds the SQL migration files so internal/migrate
// can apply them via the standard fs.FS interface regardless of the
// process's working directory.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
