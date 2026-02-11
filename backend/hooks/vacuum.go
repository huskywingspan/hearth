package hooks

import (
	"github.com/pocketbase/pocketbase"
)

// RegisterVacuum sets up a nightly VACUUM cron at 4 AM for physical data erasure.
// DELETE only marks SQLite pages as free — data remains on disk. VACUUM rewrites
// the entire database file, ensuring deleted messages are physically erased.
// This is critical for Hearth's privacy promise.
func RegisterVacuum(app *pocketbase.PocketBase) {
	app.Cron().MustAdd("hearth_nightly_vacuum", "0 4 * * *", func() {
		if _, err := app.DB().NewQuery("VACUUM").Execute(); err != nil {
			app.Logger().Error("nightly VACUUM failed", "error", err)
		} else {
			app.Logger().Info("nightly VACUUM complete")
		}
	})
}
