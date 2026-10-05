package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kyomel/audiofile/internal/interfaces"
)

type ListCommand struct {
	fs      *flag.FlagSet
	client  interfaces.Client
	baseURL string
}

func NewListCommand(client interfaces.Client, baseURL string) *ListCommand {
	return &ListCommand{
		fs:      flag.NewFlagSet("list", flag.ContinueOnError),
		client:  client,
		baseURL: baseURL,
	}
}

func (c *ListCommand) Name() string { return c.fs.Name() }

func (c *ListCommand) ParseFlags(flags []string) error { return c.fs.Parse(flags) }

func (c *ListCommand) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // ⚠️ modern
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/list", nil) // ⚠️ modern
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("list request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("list failed (%s): %s", res.Status, strings.TrimSpace(string(body)))
	}
	fmt.Println(string(body))
	return nil
}
