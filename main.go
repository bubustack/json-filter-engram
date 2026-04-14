package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	sdk "github.com/bubustack/bubu-sdk-go"
	"github.com/bubustack/json-filter-engram/pkg/engram"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := sdk.StartBatch(ctx, engram.New()); err != nil {
		log.Fatalf("json-filter engram failed: %v", err)
	}
}
