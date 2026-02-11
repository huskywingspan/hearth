package hooks

import (
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterCORS configures PocketBase's CORS policy.
// Production: AllowedOrigins locked to https://{HEARTH_DOMAIN} — never "*".
// Development: localhost:5173 (Vite dev server).
func RegisterCORS(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		domain := os.Getenv("HEARTH_DOMAIN")

		// Set the application URL for PocketBase's built-in CORS handling
		if domain != "" && domain != "localhost" && domain != "localhost:8090" {
			app.Settings().Meta.AppURL = "https://" + domain
		} else {
			app.Settings().Meta.AppURL = "http://localhost:8090"
		}

		// Add CORS middleware to reject unknown origins
		se.Router.BindFunc(func(e *core.RequestEvent) error {
			origin := e.Request.Header.Get("Origin")
			allowedOrigin := GetCORSOrigin()

			// If there's an Origin header, validate it
			if origin != "" {
				if origin != allowedOrigin {
					// Don't set any CORS headers — browser will block the response
					// For preflight OPTIONS requests, return 403
					if e.Request.Method == "OPTIONS" {
						return e.JSON(403, map[string]string{
							"error": "Origin not allowed",
						})
					}
					// For regular requests, proceed but without CORS headers
					// (browser will block reading the response)
				} else {
					// Set CORS headers for the allowed origin
					e.Response.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					e.Response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					e.Response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					e.Response.Header().Set("Access-Control-Allow-Credentials", "true")
					e.Response.Header().Set("Access-Control-Max-Age", "86400")

					// Handle preflight
					if e.Request.Method == "OPTIONS" {
						return e.String(204, "")
					}
				}
			}

			return e.Next()
		})

		return se.Next()
	})
}
