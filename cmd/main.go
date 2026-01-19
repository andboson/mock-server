package main

import (
	"log"
	"os"

	"andboson/mock-server/internal/config"
	"andboson/mock-server/internal/services/expectations"
	"andboson/mock-server/internal/services/server"
	"andboson/mock-server/internal/templates"
)

func main() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	store := expectations.NewStore()
	if err := store.AddExpectations(c.Expectations()); err != nil {
		log.Fatalf("Failed to add expectations: %v", err)
	}

	// check templates
	tpls, err := templates.NewTemplates()
	if err != nil {
		log.Fatalf("template load error: %+v", err)
	}

	srv := server.NewServer(os.Getenv(server.ServerAddrHTTP), tpls, store)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
