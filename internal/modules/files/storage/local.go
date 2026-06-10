package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/filetype"
	_ "golang.org/x/image/webp"
)

type LocalStorage struct {
	BasePath string
	BaseURL  string
}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{BasePath: basePath, BaseURL: baseURL}
}

const maxImageHeight = 300
const maxImageWidth = 300

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

	fileName := generateFileName(file.Filename)
	subDir := ""
	imageHeight := 0
	imageWidth := 0

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
	isVideo := fileKind.MIME.Type == "video"

	isAvatar := isImage && purpose == "avatar"
	isProductPicture := isImage && purpose == "product_image"
	isProductVideo := isVideo && purpose == "product_video"
	isResume := isPDF && purpose == "resume"

	if !isAvatar && !isResume && !isProductPicture && !isProductVideo {
		return "", errors.ErrUnsupported
	}

	if isAvatar {
		reader := bytes.NewReader(content)
		config, format, err := image.DecodeConfig(reader)
		if err != nil {
			return "", err
		}

		imageHeight = config.Height
		imageWidth = config.Width

		fmt.Printf("Image Format: %s\n", format)
		fmt.Printf("Width: %d pixels\n", imageWidth)
		fmt.Printf("Height: %d pixels\n", imageHeight)

		if imageHeight > maxImageHeight || imageWidth > maxImageWidth {
			return "", ErrInvalidDimensions
		}
	}

	if isAvatar {
		subDir = "avatars"
	}
	if isProductPicture {
		subDir = "products_pictures"
	}
	if isProductVideo {
		subDir = "product_videos"
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
