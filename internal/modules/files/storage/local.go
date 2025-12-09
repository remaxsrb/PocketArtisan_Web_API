package storage

import (
	"os"
	"path/filepath"
)

type LocalStorage struct {
	BasePath string
	BaseURL  string
}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{BasePath: basePath, BaseURL: baseURL}
}

func (l *LocalStorage) SaveFile(fileName string, data []byte) (string, error) {
	path := filepath.Join(l.BasePath, fileName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return l.BaseURL + "/" + fileName, nil
}

func (l *LocalStorage) DeleteFile(fileName string) error {
	path := filepath.Join(l.BasePath, fileName)
	return os.Remove(path)
}

func (l *LocalStorage) GetFileURL(fileName string) string {
	return l.BaseURL + "/" + fileName
}
