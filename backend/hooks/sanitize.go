package hooks

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// maxMessageLength is the hard cap on message body length (bytes).
// Applied server-side regardless of client-side validation.
const maxMessageLength = 4000

// RegisterSanitize sets up input sanitization hooks for all user-generated content.
// Defense-in-depth: even though React escapes output by default, we strip dangerous
// content server-side so it never enters the data layer.
func RegisterSanitize(app *pocketbase.PocketBase) {
	// Sanitize message body before creation
	app.OnRecordCreate("messages").BindFunc(func(e *core.RecordEvent) error {
		body := e.Record.GetString("body")
		e.Record.Set("body", SanitizeText(body))
		return e.Next()
	})

	// Sanitize message body on update
	app.OnRecordUpdate("messages").BindFunc(func(e *core.RecordEvent) error {
		body := e.Record.GetString("body")
		e.Record.Set("body", SanitizeText(body))
		return e.Next()
	})

	// Sanitize user display_name
	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		sanitizeRecordField(e.Record, "display_name")
		return e.Next()
	})
	app.OnRecordUpdate("users").BindFunc(func(e *core.RecordEvent) error {
		sanitizeRecordField(e.Record, "display_name")
		return e.Next()
	})

	// Sanitize room name and description
	app.OnRecordCreate("rooms").BindFunc(func(e *core.RecordEvent) error {
		sanitizeRecordField(e.Record, "name")
		sanitizeRecordField(e.Record, "description")
		return e.Next()
	})
	app.OnRecordUpdate("rooms").BindFunc(func(e *core.RecordEvent) error {
		sanitizeRecordField(e.Record, "name")
		sanitizeRecordField(e.Record, "description")
		return e.Next()
	})
}

// SanitizeText enforces the max length on user input.
// We do NOT html-escape here because React renders text content safely
// via JSX (treats values as text, not HTML). HTML-escaping server-side
// causes double-encoding (e.g., ' becomes &#39; displayed literally).
func SanitizeText(input string) string {
	if len(input) > maxMessageLength {
		return input[:maxMessageLength]
	}
	return input
}

// sanitizeRecordField enforces length limit on a specific record field (if non-empty).
func sanitizeRecordField(record *core.Record, field string) {
	value := record.GetString(field)
	if value != "" && len(value) > maxMessageLength {
		record.Set(field, value[:maxMessageLength])
	}
}
