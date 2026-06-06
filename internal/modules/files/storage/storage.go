package storage

import "mime/multipart"

type Storage interface {
	SaveFile(fileName *multipart.FileHeader, purpose string) (string, error)
	DeleteFile(fileName string) error
	GetFileURL(fileName string) string
}
