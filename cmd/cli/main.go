package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kyomel/audiofile/internal/command"
	"github.com/kyomel/audiofile/internal/interfaces"
)

const apiBase = "http://localhost:8000"

func main() {
	client := &http.Client{Timeout: 30 * time.Second} // ⚠️ modern: selalu ada timeout
	cmds := []interfaces.Command{
		command.NewGetCommand(client, apiBase),
		command.NewUploadCommand(client, apiBase),
		command.NewListCommand(client, apiBase),
	}
	if err := command.NewParser(cmds).Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err) // ⚠️ modern: Fprintf ke stderr (buku: WriteString+Sprintf)
		os.Exit(1)
	}
}
