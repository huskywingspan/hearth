package hooks

import (
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// RegisterMessageGC sets up a cron job that sweeps expired messages every minute.
// Uses the idx_messages_expires_at index for O(log n) performance.
func RegisterMessageGC(app *pocketbase.PocketBase) {
	app.Cron().MustAdd("hearth_message_gc", "* * * * *", func() {
		now := time.Now().UTC().Format(time.RFC3339)

		res, err := app.DB().
			NewQuery("DELETE FROM messages WHERE expires_at <= {:now}").
			Bind(dbx.Params{"now": now}).
			Execute()
		if err != nil {
			app.Logger().Error("message GC failed", "error", err)
			return
		}

		if affected, _ := res.RowsAffected(); affected > 0 {
			app.Logger().Info("message GC sweep", "deleted", affected)
			// Increment Prometheus counter (tracked in metrics.go)
			gcDeletedTotal.Add(affected)
		}
	})
}
