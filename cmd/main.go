package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/andboson/mock-server/internal/config"
	"github.com/andboson/mock-server/internal/services/expectations"
	"github.com/andboson/mock-server/internal/services/server"
	"github.com/andboson/mock-server/internal/templates"
)

func main() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	historyLimit := 1000
	if v := os.Getenv("HISTORY_LIMIT"); v != "" {
		historyLimit, err = strconv.Atoi(v)
		if err != nil || historyLimit < 0 {
			log.Fatalf("invalid HISTORY_LIMIT %q: must be a non-negative integer", v)
		}
	}

	store := expectations.NewStore()
	store.SetHistoryLimit(historyLimit)
	if err := store.AddExpectations(c.Expectations()); err != nil {
		log.Fatalf("Failed to add expectations: %v", err)
	}

	// check templates
	tpls, err := templates.NewTemplates()
	if err != nil {
		log.Fatalf("template load error: %+v", err)
	}

	srv := server.NewServer(os.Getenv(server.ServerAddrHTTP), tpls, store)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Stop(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
