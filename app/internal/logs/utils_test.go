package logs

import (
	"testing"
)

// TestGetCaller tests the getCaller function.
func TestGetCaller(t *testing.T) {
	// Capture the caller info from this function.
	file, line := getCaller(1)

	// Ensure that the file and line are not the placeholders.
	if file == "???" || line == 0 {
		t.Fatalf("getCaller failed to get the correct caller info. Got file: %s, line: %d", file, line)
	}

	// Test ignoring specific files
	ignoredFile := "/pkg/log/log.go"
	file, line = getCaller(1, ignoredFile)
	if file == "???" || line == 0 {
		t.Fatalf("getCaller failed to get the correct caller info after ignoring %s. Got file: %s, line: %d", ignoredFile, file, line)
	}

	// You can add more detailed checks here if needed
}

// TestGetCallerIgnoringLogMulti tests the getCallerIgnoringLogMulti function.
func TestGetCallerIgnoringLogMulti(t *testing.T) {
	// Capture the caller info from this function, which is a few levels up in the call stack.
	file, line := getCallerIgnoringLogMulti(1)

	// Ensure that the file and line are not the placeholders.
	if file == "???" || line == 0 {
		t.Fatalf("getCallerIgnoringLogMulti failed to get the correct caller info. Got file: %s, line: %d", file, line)
	}

	// Test against expected values if necessary
	// Note: The exact file and line will depend on where you place this test file.
}

// TestDecodeJSONData tests the decodeJSONData function.
func TestDecodeJSONData(t *testing.T) {
	// Define some valid JSON data.
	validJSON := []byte(`{"key1": "value1", "key2": 2}`)
	expected := map[string]interface{}{
		"key1": "value1",
		"key2": 2.0, // JSON numbers are unmarshalled as float64
	}

	// Decode the JSON data.
	result, err := decodeJSONData(validJSON)
	if err != nil {
		t.Fatalf("decodeJSONData failed with error: %v", err)
	}

	// Check if the result matches the expected map.
	for k, v := range expected {
		if rv, ok := result[k]; !ok {
			t.Errorf("key %s not found in result", k)
		} else if rv != v {
			t.Errorf("decodeJSONData(%s) = %v; want %v", string(validJSON), result, expected)
		}
	}

	// Check for any unexpected keys in the result
	for k := range result {
		if _, ok := expected[k]; !ok {
			t.Errorf("unexpected key %s in result", k)
		}
	}

	// Define some invalid JSON data.
	invalidJSON := []byte(`{"key1": "value1", "key2": 2`) // Missing closing brace.

	// Attempt to decode the invalid JSON data.
	_, err = decodeJSONData(invalidJSON)
	if err == nil {
		t.Fatal("decodeJSONData should have returned an error for invalid JSON")
	}
}
