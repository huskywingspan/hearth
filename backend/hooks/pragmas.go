package hooks

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterPragmas injects SQLite WAL pragmas on bootstrap, before any DB operations.
// These are critical for Hearth's 1GB memory budget and concurrent read performance.
func RegisterPragmas(app *pocketbase.PocketBase) {
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		// IMPORTANT: Call e.Next() FIRST to let PocketBase open the database.
		// OnBootstrap fires before the DB is initialized — we apply pragmas after.
		if err := e.Next(); err != nil {
			return err
		}

		pragmas := []string{
			"PRAGMA journal_mode=WAL",        // Non-blocking concurrent reads/writes
			"PRAGMA synchronous=NORMAL",       // Fewer fsync; sufficient for app crashes
			"PRAGMA cache_size=-2000",         // ~2MB; rely on OS filesystem cache
			"PRAGMA mmap_size=268435456",      // 256MB mmap; reduces read() syscalls
			"PRAGMA busy_timeout=5000",        // 5s lock timeout
		}

		for _, pragma := range pragmas {
			if _, err := e.App.DB().NewQuery(pragma).Execute(); err != nil {
				return fmt.Errorf("failed to set %s: %w", pragma, err)
			}
		}

		e.App.Logger().Info("SQLite PRAGMAs applied",
			"journal_mode", "WAL",
			"synchronous", "NORMAL",
			"cache_size", -2000,
			"mmap_size", 268435456,
			"busy_timeout", 5000,
		)

		return nil
	})
}
