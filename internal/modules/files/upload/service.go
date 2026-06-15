package upload

import (
	"PocketArtisan/internal/modules/files/storage"
	"mime/multipart"
)

type Service struct {
	Storage storage.Storage
}

func NewService(s storage.Storage) *Service {
	return &Service{Storage: s}
}

func (uc *Service) Execute(file *multipart.FileHeader, purpose string) (string, error) {
	return uc.Storage.SaveFile(file, purpose)
}
