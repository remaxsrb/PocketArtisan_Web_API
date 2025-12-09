package serve

import (
	"PocketArtisan/internal/modules/files/storage"
	"fmt"
)

type UseCase struct {
	Storage storage.Storage
}

func NewUseCase(s storage.Storage) *UseCase {
	return &UseCase{Storage: s}
}

func (uc *UseCase) Execute(filename string) (string, error) {
	url := uc.Storage.GetFileURL(filename)
	if url == "" {
		return "", fmt.Errorf("file not found")
	}
	return url, nil
}
