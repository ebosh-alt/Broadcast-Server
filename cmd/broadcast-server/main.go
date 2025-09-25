package main

import (
	"flag"
	"log"

	"broadcast_server/internal/app"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "HTTP listen port")
	flag.Parse()

	log.Printf("broadcast-server %s (%s) built %s", version, commit, buildDate)

	if err := app.Run(app.Config{Addr: ":" + port}); err != nil {
		log.Fatal(err)
	}
}
