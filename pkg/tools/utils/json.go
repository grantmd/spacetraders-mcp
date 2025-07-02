package utils

import (
	"encoding/json"
)

// FormatJSON formats a data structure as properly indented JSON string
// Returns a formatted JSON string or an error message if marshaling fails
func FormatJSON(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// Return error message if JSON marshaling fails
		return "Error formatting JSON: " + err.Error()
	}
	return string(jsonBytes)
}
