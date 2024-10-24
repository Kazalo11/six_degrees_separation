package main

import (
	"log"

	"github.com/Kazalo11/six-degrees-seperation/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	server.Start()
}
