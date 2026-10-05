package command

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyomel/audiofile/internal/interfaces"
)

type UploadCommand struct {
	fs       *flag.FlagSet
	client   interfaces.Client
	baseURL  string
	filename string
}

func NewUploadCommand(client interfaces.Client, baseURL string) *UploadCommand {
	cmd := &UploadCommand{
		fs:      flag.NewFlagSet("upload", flag.ContinueOnError),
		client:  client,
		baseURL: baseURL,
	}
	cmd.fs.StringVar(&cmd.filename, "filename", "", "full path of the file to upload")
	return cmd
}

func (c *UploadCommand) Name() string { return c.fs.Name() }

func (c *UploadCommand) ParseFlags(flags []string) error {
	if len(flags) == 0 {
		fmt.Println("usage: audiofile-cli upload -filename <file>")
		return errors.New("missing flags")
	}
	return c.fs.Parse(flags)
}

func (c *UploadCommand) Run() error {
	if c.filename == "" {
		return errors.New("missing filename")
	}
	fmt.Println("Uploading", c.filename, "...")

	file, err := os.Open(c.filename)
	if err != nil {
		return fmt.Errorf("open %s: %w", c.filename, err) // ⚠️ modern: %w biar bisa errors.Is/As
	}
	defer file.Close()

	var payload bytes.Buffer
	mw := multipart.NewWriter(&payload)
	part, err := mw.CreateFormFile("file", filepath.Base(c.filename))
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("close multipart: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // ⚠️ modern
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/upload", &payload) // ⚠️ modern
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("upload request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if res.StatusCode != http.StatusOK { // ⚠️ modern: buku nggak cek status code
		return fmt.Errorf("upload failed (%s): %s", res.Status, strings.TrimSpace(string(body)))
	}
	fmt.Println("Audiofile ID:", string(body))
	return nil
}
