package serve

import (
	"PocketArtisan/internal/modules/files/storage"
	"fmt"
)

type Service struct {
	Storage storage.Storage
}

func NewService(s storage.Storage) *Service {
	return &Service{Storage: s}
}

func (uc *Service) Execute(filename string) (string, error) {
	url := uc.Storage.GetFileURL(filename)
	if url == "" {
		return "", fmt.Errorf("file not found")
	}
	return url, nil
}
