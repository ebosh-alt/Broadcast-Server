package main

import (
	"broadcast_server/internal/app"
	"flag"
	"log"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "8080", "The port to listen on")
	flag.Parse()
	log.Printf("Starting server on port :%s", port)

	if err := app.Run(app.Config{Addr: ":" + port}); err != nil {
		log.Fatal(err)
	}
}
