package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"collector/go/internal/collector"
)

func main() {
	symbol := flag.String("symbol", "BTCUSDT", "Trading pair symbol to fetch")
	outputDir := flag.String("output", "data", "Directory where NDJSON files are stored")
	interval := flag.Duration("interval", time.Second, "Interval between fetches (e.g. 1s)")
	flag.Parse()

	if *interval <= 0 {
		log.Fatalf("interval must be greater than zero")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := &http.Client{Timeout: 5 * time.Second}

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	log.Printf("Starting collector for %s every %s", *symbol, *interval)

	if err := runOnce(ctx, client, *outputDir, *symbol); err != nil {
		log.Printf("collect failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down collector")
			return
		case <-ticker.C:
			if err := runOnce(ctx, client, *outputDir, *symbol); err != nil {
				log.Printf("collect failed: %v", err)
			}
		}
	}
}

func runOnce(ctx context.Context, client *http.Client, outputDir, symbol string) error {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filePath, err := collector.Collect(reqCtx, client, outputDir, symbol, time.Now())
	if err != nil {
		return err
	}

	log.Printf("Snapshot stored in %s", filePath)
	return nil
}
