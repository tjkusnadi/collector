package collector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client knows how to pull raw payloads from Binance and persist them.
type Client struct {
	httpClient *http.Client
	endpoint   string
	outDir     string
	symbol     string
}

// NewClient creates a collector client.
func NewClient(httpClient *http.Client, endpoint, symbol, outDir string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
		outDir:     outDir,
		symbol:     symbol,
	}
}

// FetchAndStore downloads the payload and appends the raw response body to disk.
func (c *Client) FetchAndStore(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if err := os.MkdirAll(c.outDir, 0o755); err != nil {
		return "", fmt.Errorf("ensure output dir: %w", err)
	}

	now := time.Now().UTC()
	filename := fmt.Sprintf("%s-%s.ndjson", c.symbol, now.Format("2006-01-02-15"))
	path := filepath.Join(c.outDir, filename)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("open payload file: %w", err)
	}
	defer file.Close()

	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	if _, err := file.Write(data); err != nil {
		return "", fmt.Errorf("append payload: %w", err)
	}

	return path, nil
}

// Run continuously fetches data at the supplied interval until the context is cancelled.
func (c *Client) Run(ctx context.Context, interval time.Duration, onTick func(string)) error {
	if interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		path, err := c.FetchAndStore(ctx)
		if err != nil {
			return err
		}
		if onTick != nil {
			onTick(path)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
