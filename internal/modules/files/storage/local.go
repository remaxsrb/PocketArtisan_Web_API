package storage

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalStorage struct {
	BasePath string
	BaseURL  string
}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{BasePath: basePath, BaseURL: baseURL}
}

const maxImageHeight = 500
const maxImageWidth = 500

var (
	ErrInvalidDimensions = errors.New("Maksimalne dimenzije za profilnu sliku su 300x300 piksela")
	ErrUnsupportedFormat = errors.New("nepodržan format fajla ili namena")
)

func generateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)

	random := make([]byte, 4)
	rand.Read(random)

	return fmt.Sprintf(
		"%s_%d_%s%s",
		name,
		time.Now().UnixMilli(),
		hex.EncodeToString(random),
		ext,
	)
}

func (l *LocalStorage) SaveFile(file *multipart.FileHeader, purpose string) (string, error) {
	fileName, subDir, content, err := resolveUpload(file, purpose)
	if err != nil {
		return "", err
	}

	path := filepath.Join(l.BasePath, subDir, fileName)
	os.MkdirAll(filepath.Dir(path), 0755)

	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", err
	}
	return l.BaseURL + "/" + subDir + "/" + fileName, nil
}

func (l *LocalStorage) SaveRawFile(data []byte, filename, subDir string) (string, error) {
	path := filepath.Join(l.BasePath, subDir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return l.BaseURL + "/" + subDir + "/" + filename, nil
}

func (l *LocalStorage) DeleteFile(fileName string) error {
	path := filepath.Join(l.BasePath, fileName)
	return os.Remove(path)
}

func (l *LocalStorage) GetFileURL(fileName string) string {
	return l.BaseURL + "/" + fileName
}
