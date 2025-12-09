package storage

type Storage interface {
	SaveFile(fileName string, data []byte) (string, error)
	DeleteFile(fileName string) error
	GetFileURL(fileName string) string
}
