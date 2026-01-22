package migrate

import "embed"

//go:embed migrations/*/*.sql
var EmbeddedFS embed.FS
