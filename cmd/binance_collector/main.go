package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"collector/internal/collector"
)

const defaultEndpointTemplate = "https://www.binance.com/bapi/asset/v2/public/asset-service/product/get-product-by-symbol?symbol=%s"

func main() {
	var (
		symbol      = flag.String("symbol", "BTCUSDT", "Binance trading pair symbol")
		outDir      = flag.String("out", filepath.Join("data", "raw"), "directory for raw payloads")
		intervalSec = flag.Int("interval", 1, "poll interval in seconds")
	)
	flag.Parse()

	endpoint := fmt.Sprintf(defaultEndpointTemplate, *symbol)
	interval := time.Duration(*intervalSec) * time.Second

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("starting binance collector: symbol=%s interval=%s output=%s", *symbol, interval, *outDir)

	client := collector.NewClient(&http.Client{Timeout: 10 * time.Second}, endpoint, *outDir)

	if err := client.Run(ctx, interval, func(path string) {
		log.Printf("stored payload at %s", path)
	}); err != nil {
		if ctx.Err() != nil {
			log.Printf("collector stopped: %v", ctx.Err())
		} else {
			log.Fatalf("collector error: %v", err)
		}
	}
}
