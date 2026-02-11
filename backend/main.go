package main

import (
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"

	"hearth/hooks"
)

func main() {
	app := pocketbase.New()

	// Phase 1: Data Layer
	hooks.RegisterPragmas(app)
	hooks.RegisterCollections(app)
	hooks.RegisterMessageGC(app)
	hooks.RegisterVacuum(app)
	hooks.RegisterPresence(app)

	// Phase 2: Auth & Security
	hooks.RegisterAuth(app)
	hooks.RegisterInvite(app)
	hooks.RegisterPoW(app)
	hooks.RegisterLiveKitToken(app)
	hooks.RegisterRateLimit(app)
	hooks.RegisterSanitize(app)
	hooks.RegisterCORS(app)

	// Phase 3: Observability
	hooks.RegisterMetrics(app)

	// Serve the Hearth SPA from pb_public/ (with SPA index fallback).
	// PocketBase only auto-serves pb_public when using the prebuilt binary;
	// when used as a Go framework, we must register it explicitly.
	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			if !e.Router.HasRoute(http.MethodGet, "/{path...}") {
				e.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), true))
			}
			return e.Next()
		},
		Priority: 999, // run last so custom API routes take precedence
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
