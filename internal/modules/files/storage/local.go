package storage

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/h2non/filetype"
)

type LocalStorage struct {
	BasePath string
	BaseURL  string
}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{BasePath: basePath, BaseURL: baseURL}
}

func (l *LocalStorage) SaveFile(file *multipart.FileHeader, purpose string) (string, error) {

	fileName := file.Filename
	subDir := ""

	src, _ := file.Open()
	defer src.Close()
	content, _ := io.ReadAll(src)

	head := make([]byte, 8192)
	if len(content) > 8192 {
		head = content[:8192]
	} else {
		head = content
	}

	fileKind, _ := filetype.Match(head)

	isImage := fileKind.MIME.Type == "image"
	isPDF := fileKind.Extension == "pdf"

	isAvatar := isImage && purpose == "avatar"
	isProduct := isImage && purpose == "product"
	isResume := isPDF && purpose == "resume"

	if !isAvatar && !isResume && !isProduct {
		return "", errors.ErrUnsupported
	}

	if isAvatar {
		subDir = "avatars"
	}
	if isProduct {
		subDir = "products"
	}
	if isResume {
		subDir = "resumes"
	}

	path := filepath.Join(l.BasePath, subDir, fileName)
	os.MkdirAll(filepath.Dir(path), 0755)

	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", err
	}
	return l.BaseURL + "/" + subDir + "/" + fileName, nil
}

func (l *LocalStorage) DeleteFile(fileName string) error {
	path := filepath.Join(l.BasePath, fileName)
	return os.Remove(path)
}

func (l *LocalStorage) GetFileURL(fileName string) string {
	return l.BaseURL + "/" + fileName
}
