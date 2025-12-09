package upload

import "PocketArtisan/internal/modules/files/storage"

type UseCase struct {
	Storage storage.Storage
}

func NewUseCase(s storage.Storage) *UseCase {
	return &UseCase{Storage: s}
}

func (uc *UseCase) Execute(filename string, content []byte) (string, error) {
	return uc.Storage.SaveFile(filename, content)
}
