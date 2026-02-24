package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"rpg-game/pkg/auth"
	"rpg-game/pkg/db"
	"rpg-game/pkg/server"
)

func main() {
	dbPath := flag.String("db", "game.db", "path to SQLite database file")
	addr := flag.String("addr", ":8080", "listen address")
	secret := flag.String("secret", "change-me-in-production", "JWT signing secret")
	staticDir := flag.String("static", "web/static", "path to static files directory")
	flag.Parse()

	store, err := db.NewStore(*dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	authService := auth.NewAuthService(store, *secret)
	srv := server.NewServer(store, authService, *staticDir)

	httpServer := &http.Server{
		Addr:         *addr,
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("RPG game server starting on %s\n", *addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
