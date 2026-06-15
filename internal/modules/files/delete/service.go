package delete

import "PocketArtisan/internal/modules/files/storage"

type Service struct {
	Storage storage.Storage
}

func NewService(s storage.Storage) *Service {
	return &Service{Storage: s}
}

func (uc *Service) Execute(filename string) error {
	return uc.Storage.DeleteFile(filename)
}
