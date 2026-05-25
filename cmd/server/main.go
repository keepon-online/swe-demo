package main

import (
	"log"

	"github.com/keepon-online/swe-demo/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
