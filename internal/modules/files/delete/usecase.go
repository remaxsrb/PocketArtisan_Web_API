package delete

import "PocketArtisan/internal/modules/files/storage"

type UseCase struct {
	Storage storage.Storage
}

func NewUseCase(s storage.Storage) *UseCase {
	return &UseCase{Storage: s}
}

func (uc *UseCase) Execute(filename string) error {
	return uc.Storage.DeleteFile(filename)
}
