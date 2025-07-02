package utils

import (
	"testing"
)

func TestFormatJSON(t *testing.T) {
	// Test with a simple map
	data := map[string]interface{}{
		"name":  "test",
		"value": 123,
		"items": []string{"a", "b", "c"},
	}

	result := FormatJSON(data)

	// Check that result contains properly formatted JSON
	if !contains(result, `"name": "test"`) {
		t.Errorf("Expected JSON to contain formatted name field, got: %s", result)
	}

	if !contains(result, `"value": 123`) {
		t.Errorf("Expected JSON to contain formatted value field, got: %s", result)
	}

	if !contains(result, `"items": [`) {
		t.Errorf("Expected JSON to contain formatted items array, got: %s", result)
	}

	// Check that it's properly indented (should contain newlines and spaces)
	if !contains(result, "\n") {
		t.Errorf("Expected indented JSON with newlines, got: %s", result)
	}
}

func TestFormatJSON_EmptyMap(t *testing.T) {
	data := map[string]interface{}{}
	result := FormatJSON(data)

	expected := "{}"
	if result != expected {
		t.Errorf("Expected empty object JSON '%s', got: %s", expected, result)
	}
}

func TestFormatJSON_Nil(t *testing.T) {
	result := FormatJSON(nil)

	expected := "null"
	if result != expected {
		t.Errorf("Expected 'null' for nil input, got: %s", result)
	}
}

func TestFormatJSON_ComplexStructure(t *testing.T) {
	data := map[string]interface{}{
		"ship": map[string]interface{}{
			"symbol": "SHIP_1234",
			"nav": map[string]interface{}{
				"status":   "IN_ORBIT",
				"system":   "X1-TEST",
				"waypoint": "X1-TEST-A1",
				"coordinates": map[string]interface{}{
					"x": 10,
					"y": 20,
				},
			},
		},
		"success": true,
	}

	result := FormatJSON(data)

	// Check nested structure formatting
	if !contains(result, `"ship": {`) {
		t.Errorf("Expected nested ship object, got: %s", result)
	}

	if !contains(result, `"nav": {`) {
		t.Errorf("Expected nested nav object, got: %s", result)
	}

	if !contains(result, `"coordinates": {`) {
		t.Errorf("Expected nested coordinates object, got: %s", result)
	}

	if !contains(result, `"success": true`) {
		t.Errorf("Expected boolean value formatting, got: %s", result)
	}
}

func TestFormatJSON_InvalidData(t *testing.T) {
	// Use a channel, which cannot be marshaled to JSON
	data := make(chan int)
	result := FormatJSON(data)

	if !contains(result, "Error formatting JSON") {
		t.Errorf("Expected error message for unmarshalable data, got: %s", result)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
