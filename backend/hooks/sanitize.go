package hooks

import (
	"html"

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

// SanitizeText HTML-escapes text and enforces the max length.
// Turns <script>alert(1)</script> into &lt;script&gt;alert(1)&lt;/script&gt;
// Uses Go's html.EscapeString (standard library, handles all HTML entities correctly).
func SanitizeText(input string) string {
	sanitized := html.EscapeString(input)
	if len(sanitized) > maxMessageLength {
		sanitized = sanitized[:maxMessageLength]
	}
	return sanitized
}

// sanitizeRecordField escapes HTML in a specific record field (if non-empty).
func sanitizeRecordField(record *core.Record, field string) {
	value := record.GetString(field)
	if value != "" {
		record.Set(field, html.EscapeString(value))
	}
}
