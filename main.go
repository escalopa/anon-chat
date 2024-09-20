package main

import (
	"os"

	log "github.com/catalystgo/logger/cli"
	"github.com/escalopa/anon-chat-app/room"
	"github.com/escalopa/anon-chat-app/server"
	"github.com/escalopa/anon-chat-app/storage"
)

var (
	port     string = "8080"      // HTTP server port (default: 8080) (overwritten by env var PORT)
	dataPath string = "data.json" // Data file path (default: data.json) (overwritten by env var DATA_PATH)
)

func loadEnv() {
	if v := os.Getenv("PORT"); v != "" {
		log.Debugf("Got port from env: %s", v)
		port = v
	}

	if v := os.Getenv("DATA_PATH"); v != "" {
		log.Debugf("Got data file from env: %s", v)
		dataPath = v
	}
}

func main() {
	loadEnv()
	log.SetLevel(log.LevelInfo)

	log.Infof("Using port: %s", port)
	log.Infof("Using data file: %s", dataPath)

	// Create a new storage and load the data
	db := storage.New()
	if err := db.Load(dataPath); err != nil {
		log.Fatalf("Error loading data: %v", err)
	}
	log.Infof("Loaded %d messages", db.Count())

	// Create a new chat server and run it
	chatServer := room.New(db)
	go chatServer.Run()

	defer func() {
		// Dump the data before exiting
		if errDump := db.Dump(dataPath); errDump != nil {
			log.Errorf("Error dumping data: %v", errDump)
		}
		log.Infof("Data dumped")
	}()

	// Create a new HTTP server
	srv := server.New(port, db, chatServer)

	// Run the HTTP server
	if err := srv.Run(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
