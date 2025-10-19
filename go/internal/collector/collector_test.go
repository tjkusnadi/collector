package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type stubClient struct {
	status int
	body   []byte
}

func (s *stubClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: s.status,
		Body:       io.NopCloser(bytes.NewReader(s.body)),
		Header:     make(http.Header),
	}, nil
}

func TestCollectWritesNDJSONWithRotation(t *testing.T) {
	payload := map[string]any{"symbol": "BTCUSDT", "price": "12345.67"}
	body, err := json.Marshal(map[string]any{"data": payload})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	client := &stubClient{status: http.StatusOK, body: body}

	dir := t.TempDir()
	now := time.Date(2024, 3, 1, 15, 30, 0, 0, time.UTC)

	filePath, err := Collect(context.Background(), client, dir, "btcusdt", now)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	expectedName := filepath.Join(dir, "BTCUSDT-2024030115.ndjson")
	if filePath != expectedName {
		t.Fatalf("unexpected file path: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	lines := bytesSplit(data, '\n')
	if len(lines) != 1 {
		t.Fatalf("expected one JSON line, got %q", lines)
	}

	var result Result
	if err := json.Unmarshal(lines[0], &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected symbol: %s", result.Symbol)
	}

	if !result.CollectedAt.Equal(now.UTC()) {
		t.Fatalf("unexpected timestamp: %s", result.CollectedAt)
	}

	if result.Payload["price"] != payload["price"] {
		t.Fatalf("unexpected payload: %#v", result.Payload)
	}
}

func bytesSplit(b []byte, sep byte) [][]byte {
	var parts [][]byte
	start := 0
	for i, c := range b {
		if c == sep {
			parts = append(parts, b[start:i])
			start = i + 1
		}
	}
	if start < len(b) {
		parts = append(parts, b[start:])
	}
	return parts
}
