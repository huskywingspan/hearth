package main

import (
	"log"

	"github.com/pocketbase/pocketbase"

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

	// Phase 3: Observability
	hooks.RegisterMetrics(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
