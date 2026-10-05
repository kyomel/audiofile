package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/kyomel/audiofile/internal/interfaces"
	"github.com/kyomel/audiofile/internal/storage"
	"github.com/kyomel/audiofile/models"
)

type MetadataService struct {
	Storage interfaces.Storage
	Server  *http.Server
}

func CreateMetadataService(port int, store interfaces.Storage) *MetadataService {
	mux := http.NewServeMux()
	ms := &MetadataService{Storage: store}

	mux.HandleFunc("POST /upload", ms.uploadHandler)
	mux.HandleFunc("GET /request", ms.getByIDHandler)
	mux.HandleFunc("GET /list", ms.listHandler)

	ms.Server = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return ms
}

func Run(port int) error {
	service := CreateMetadataService(port, &storage.FlatFile{})
	slog.Info("metadata service listening", "addr", service.Server.Addr)
	return service.Server.ListenAndServe()
}

func (ms *MetadataService) uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read upload", http.StatusInternalServerError)
		return
	}

	id, path, err := ms.Storage.Upload(data, header.Filename)
	if err != nil {
		slog.Error("upload failed", "err", err)
		http.Error(w, "store upload", http.StatusInternalServerError)
		return
	}

	audio := &models.Audio{Id: id, Path: path, Status: "uploaded"}
	if err := ms.Storage.SaveMetadata(audio); err != nil {
		slog.Error("save metadata failed", "err", err)
		http.Error(w, "save metadata", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(id))
}

func (ms *MetadataService) getByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	audio, err := ms.Storage.GetById(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "audio not found", http.StatusNotFound)
			return
		}
		slog.Error("get failed", "err", err)
		http.Error(w, "get audio", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(audio)
}

func (ms *MetadataService) listHandler(w http.ResponseWriter, r *http.Request) {
	audios, err := ms.Storage.List()
	if err != nil {
		slog.Error("list failed", "err", err)
		http.Error(w, "list audio", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(audios)
}
