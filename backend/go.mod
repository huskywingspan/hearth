module hearth

go 1.23.0

// Run `go mod tidy` after cloning to resolve all transitive dependencies.
// PocketBase v0.36+ uses pure-Go SQLite (modernc.org/sqlite) — no CGo needed.
require (
	github.com/livekit/protocol v1.24.0
	github.com/pocketbase/dbx v1.11.0
	github.com/pocketbase/pocketbase v0.26.6
)
