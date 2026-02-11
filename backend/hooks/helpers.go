package hooks

import (
	"github.com/pocketbase/dbx"
)

// dbxParams creates a dbx.Params map from alternating key-value pairs.
// Usage: dbxParams("room", roomId, "user", userId)
func dbxParams(pairs ...string) dbx.Params {
	params := dbx.Params{}
	for i := 0; i < len(pairs)-1; i += 2 {
		params[pairs[i]] = pairs[i+1]
	}
	return params
}
