package main

import (
	"log"
	"os"
	"path/filepath"
	"test-auth/internal/config"
	"test-auth/internal/server"

	"github.com/joho/godotenv"
)

func init() {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fullPath := filepath.Join(filepath.Join(path, "../.."), ".env")
	err = godotenv.Load(fullPath)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	cfg := config.MustLoad()

	s := server.New(cfg)
	err := s.Run(cfg.Port)
	if err != nil {
		panic(err)
	}
}
