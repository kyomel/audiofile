package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kyomel/audiofile/models"
)

var ErrNotFound = errors.New("not found")

// FlatFile stores audio bytes under data/ and metadata in one JSON file.
// Delete removes the whole entry; the tag argument is unused for now.
type FlatFile struct {
	mu sync.Mutex
}

func (f *FlatFile) Upload(data []byte, filename string) (string, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate id: %w", err)
	}
	id := hex.EncodeToString(buf)

	if err := os.MkdirAll("data", 0o755); err != nil {
		return "", "", fmt.Errorf("create data dir: %w", err)
	}
	path := filepath.Join("data", id+"-"+filepath.Base(filename))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", "", fmt.Errorf("write file: %w", err)
	}
	return id, path, nil
}

func (f *FlatFile) SaveMetadata(audio *models.Audio) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	all, err := f.readAll()
	if err != nil {
		return err
	}
	return f.writeAll(append(all, audio))
}

func (f *FlatFile) List() ([]*models.Audio, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.readAll()
}

func (f *FlatFile) GetById(id string) (*models.Audio, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	all, err := f.readAll()
	if err != nil {
		return nil, err
	}
	for _, audio := range all {
		if audio.Id == id {
			return audio, nil
		}
	}
	return nil, fmt.Errorf("audio %s: %w", id, ErrNotFound)
}

func (f *FlatFile) Delete(id, tag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	all, err := f.readAll()
	if err != nil {
		return err
	}

	var target *models.Audio
	kept := make([]*models.Audio, 0, len(all))
	for _, audio := range all {
		if audio.Id == id {
			target = audio
			continue
		}
		kept = append(kept, audio)
	}
	if target == nil {
		return fmt.Errorf("audio %s: %w", id, ErrNotFound)
	}
	if err := f.writeAll(kept); err != nil {
		return err
	}
	return os.Remove(target.Path)
}

func (f *FlatFile) metadataPath() string { return filepath.Join("data", "metadata.json") }

func (f *FlatFile) readAll() ([]*models.Audio, error) {
	raw, err := os.ReadFile(f.metadataPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read metadata: %w", err)
	}
	var all []*models.Audio
	if err := json.Unmarshal(raw, &all); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	return all, nil
}

func (f *FlatFile) writeAll(all []*models.Audio) error {
	raw, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	if err := os.MkdirAll("data", 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	if err := os.WriteFile(f.metadataPath(), raw, 0o644); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}
	return nil
}
