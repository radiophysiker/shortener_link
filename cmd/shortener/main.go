package main

import (
	"log"

	"github.com/radiophysiker/shortener_link/internal/app"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatalf("cannot run the app! %v", err)
	}
}
