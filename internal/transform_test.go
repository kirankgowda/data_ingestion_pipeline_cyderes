package internal

import (
	"testing"
)

//To test the transformData function with some mocked values.
func TestTransformData(t *testing.T) {
	input := []map[string]interface{}{
		{"id": 1, "title": "test"},
	}

	transformed, err := transformData(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(transformed) != 1 {
		t.Errorf("expected 1 item, got %d", len(transformed))
	}

	ingestedAt, ok := transformed[0]["ingested_at"].(string)
	if !ok || ingestedAt == "" {
		t.Error("missing or invalid 'ingested_at' field")
	}

	if transformed[0]["source"] != "placeholder_api" {
		t.Error("expected source to be 'placeholder_api'")
	}
}
