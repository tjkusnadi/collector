package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HTTPClient defines the minimal interface required from an HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Result represents a single collected snapshot.
type Result struct {
	Symbol      string                 `json:"symbol"`
	CollectedAt time.Time              `json:"collected_at"`
	Payload     map[string]interface{} `json:"payload"`
}

// Collect fetches the Binance product information for the given symbol and
// appends it as a NDJSON entry with hourly rotation.
func Collect(ctx context.Context, client HTTPClient, baseDir, symbol string, now time.Time) (string, error) {
	if client == nil {
		return "", errors.New("client is required")
	}

	if symbol == "" {
		return "", errors.New("symbol is required")
	}

	if baseDir == "" {
		return "", errors.New("baseDir is required")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://www.binance.com/bapi/asset/v2/public/asset-service/product/get-product-by-symbol?symbol=%s", symbol),
		nil,
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "collector-go/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var raw struct {
		Data map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return "", err
	}

	entry := Result{
		Symbol:      strings.ToUpper(symbol),
		CollectedAt: now.UTC(),
		Payload:     raw.Data,
	}

	if entry.Payload == nil {
		entry.Payload = map[string]interface{}{}
	}

	outDir := filepath.Clean(baseDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%s-%s.ndjson", strings.ToUpper(symbol), now.UTC().Format("2006010215"))
	filePath := filepath.Join(outDir, fileName)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}

	if _, err := f.Write(append(jsonEntry, '\n')); err != nil {
		return "", err
	}

	return filePath, nil
}
