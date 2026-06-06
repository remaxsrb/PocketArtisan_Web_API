package storage

import (
	"io"
	"mime/multipart"
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

func (l *LocalStorage) SaveFile(file *multipart.FileHeader, purpose string) (string, error) {

	src, _ := file.Open()
	defer src.Close()
	content, _ := io.ReadAll(src)

	//head := make([]byte, 8192)
	//if len(content) > 8192 {
	//	head = content[:8192]
	//} else {
	//	head = content
	//}

	fileName := file.Filename
	path := filepath.Join(l.BasePath, purpose, fileName)
	os.MkdirAll(filepath.Dir(path), 0755)

	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", err
	}
	return l.BaseURL + "/" + purpose + "/" + fileName, nil
}

func (l *LocalStorage) DeleteFile(fileName string) error {
	path := filepath.Join(l.BasePath, fileName)
	return os.Remove(path)
}

func (l *LocalStorage) GetFileURL(fileName string) string {
	return l.BaseURL + "/" + fileName
}
