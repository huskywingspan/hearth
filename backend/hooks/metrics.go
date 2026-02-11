package hooks

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// gcDeletedTotal tracks the cumulative count of messages deleted by GC.
// Exported as a Prometheus counter at /metrics.
var gcDeletedTotal atomic.Int64

// RegisterMetrics exposes a Prometheus-compatible /metrics endpoint.
// Metrics: Go heap, goroutines, room count, online users, messages, GC deletes, WAL pages.
func RegisterMetrics(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/metrics", func(e *core.RequestEvent) error {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			var b strings.Builder

			// Go runtime metrics
			writeGauge(&b, "hearth_go_heap_bytes", "Current Go heap allocation in bytes", float64(memStats.HeapAlloc))
			writeGauge(&b, "hearth_go_heap_sys_bytes", "Total bytes of heap obtained from OS", float64(memStats.HeapSys))
			writeGauge(&b, "hearth_go_goroutines", "Current number of goroutines", float64(runtime.NumGoroutine()))
			writeGauge(&b, "hearth_go_gc_runs_total", "Total number of completed GC cycles", float64(memStats.NumGC))

			// Application metrics
			roomCount := countRecords(e.App, "rooms")
			writeGauge(&b, "hearth_rooms_total", "Total number of rooms", float64(roomCount))

			msgCount := countRecords(e.App, "messages")
			writeGauge(&b, "hearth_messages_total", "Total number of active messages", float64(msgCount))

			userCount := countRecords(e.App, "users")
			writeGauge(&b, "hearth_users_total", "Total registered users", float64(userCount))

			onlineCount := presence.OnlineCount()
			writeGauge(&b, "hearth_users_online", "Currently online users", float64(onlineCount))

			// GC metrics
			writeCounter(&b, "hearth_gc_deleted_total", "Total messages deleted by GC", float64(gcDeletedTotal.Load()))

			// SQLite WAL metrics
			walPages, checkpointedPages := getWALStats(e.App)
			writeGauge(&b, "hearth_sqlite_wal_pages", "Current WAL log pages", float64(walPages))
			writeGauge(&b, "hearth_sqlite_wal_checkpointed", "WAL pages already checkpointed", float64(checkpointedPages))

			e.Response.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
			return e.String(200, b.String())
		})

		return se.Next()
	})
}

// writeGauge writes a Prometheus gauge metric to the builder.
func writeGauge(b *strings.Builder, name, help string, value float64) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s gauge\n", name)
	fmt.Fprintf(b, "%s %g\n", name, value)
}

// writeCounter writes a Prometheus counter metric to the builder.
func writeCounter(b *strings.Builder, name, help string, value float64) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s counter\n", name)
	fmt.Fprintf(b, "%s %g\n", name, value)
}

// countRecords returns the count of records in a collection, or 0 on error.
func countRecords(app core.App, collection string) int {
	total, err := app.CountRecords(collection)
	if err != nil {
		return 0
	}
	return total
}

// getWALStats queries SQLite WAL checkpoint status.
func getWALStats(app core.App) (walPages int, checkpointedPages int) {
	type walResult struct {
		Busy        int `db:"busy"`
		Log         int `db:"log"`
		Checkpointed int `db:"checkpointed"`
	}

	var result walResult
	err := app.DB().
		NewQuery("PRAGMA wal_checkpoint(PASSIVE)").
		One(&result)
	if err != nil {
		return 0, 0
	}
	return result.Log, result.Checkpointed
}
