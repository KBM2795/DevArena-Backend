package main

import (
	"log"

	"github.com/KBM2795/DevArena-Backend/internal/config"
	"github.com/KBM2795/DevArena-Backend/internal/database"
	"github.com/KBM2795/DevArena-Backend/internal/server"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to Database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Initialize and Start Server
	srv := server.NewServer(cfg.Server, db, cfg.Env)
	if err := srv.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
