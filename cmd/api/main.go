package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/kyomel/audiofile/services/metadata"
)

func main() {
	port := flag.Int("p", 8000, "port for the metadata service")
	flag.Parse()
	if err := metadata.Run(*port); err != nil {
		slog.Error("api stopped", "err", err)
		os.Exit(1)
	}
}
