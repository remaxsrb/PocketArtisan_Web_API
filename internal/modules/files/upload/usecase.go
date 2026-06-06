package upload

import (
	"PocketArtisan/internal/modules/files/storage"
	"mime/multipart"
)

type UseCase struct {
	Storage storage.Storage
}

func NewUseCase(s storage.Storage) *UseCase {
	return &UseCase{Storage: s}
}

func (uc *UseCase) Execute(file *multipart.FileHeader, purpose string) (string, error) {
	return uc.Storage.SaveFile(file, purpose)
}
