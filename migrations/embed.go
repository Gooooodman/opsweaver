// Package migrations exposes embedded SQL migrations for OpsWeaver databases.
package migrations

import "embed"

//go:embed opsweaver_server/*.sql opsweaver_gateway/*.sql
var FS embed.FS
