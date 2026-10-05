package interfaces

import (
	"net/http"

	"github.com/kyomel/audiofile/models"
)

type Command interface {
	Name() string
	ParseFlags([]string) error
	Run() error
}

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type Storage interface {
	Upload(data []byte, filename string) (id string, path string, err error)
	SaveMetadata(audio *models.Audio) error
	List() ([]*models.Audio, error)
	GetById(id string) (*models.Audio, error)
	Delete(id, tag string) error
}
