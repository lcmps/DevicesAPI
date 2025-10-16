package main

import (
	"log"

	"github.com/lcmps/DevicesAPI/db"
	"github.com/lcmps/DevicesAPI/web"
)

func main() {
	database, err := db.New()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	if err := database.Init(); err != nil {
		log.Fatalf("failed to run migrations/init: %v", err)
	}

	web.New(database).Serve()
}
