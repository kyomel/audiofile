package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyomel/audiofile/internal/interfaces"
)

type GetCommand struct {
	fs      *flag.FlagSet
	client  interfaces.Client
	baseURL string
	id      string
}

func NewGetCommand(client interfaces.Client, baseURL string) *GetCommand {
	cmd := &GetCommand{
		fs:      flag.NewFlagSet("get", flag.ContinueOnError),
		client:  client,
		baseURL: baseURL,
	}
	cmd.fs.StringVar(&cmd.id, "id", "", "id of the audio file")
	return cmd
}

func (c *GetCommand) Name() string { return c.fs.Name() }

func (c *GetCommand) ParseFlags(flags []string) error {
	if len(flags) == 0 {
		fmt.Println("usage: audiofile-cli get -id <id>")
		return errors.New("missing flags")
	}
	return c.fs.Parse(flags)
}

func (c *GetCommand) Run() error {
	if c.id == "" {
		return errors.New("missing id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // ⚠️ modern
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/request?id="+url.QueryEscape(c.id), nil) // ⚠️ modern
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("get failed (%s): %s", res.Status, strings.TrimSpace(string(body)))
	}
	fmt.Println(string(body))
	return nil
}
