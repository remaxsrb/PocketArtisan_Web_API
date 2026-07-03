package storage

import (
	"bytes"
	"errors"
	"image"
	"io"
	"mime/multipart"

	_ "image/jpeg"
	_ "image/png"

	"github.com/h2non/filetype"
	_ "golang.org/x/image/webp"
)

// resolveUpload validates a multipart file against the requested purpose and
// returns the generated file name, its target sub-directory and the raw file
// content. It centralizes the validation logic shared by every Storage
// implementation.
func resolveUpload(file *multipart.FileHeader, purpose string) (fileName, subDir string, content []byte, err error) {
	fileName = generateFileName(file.Filename)

	src, err := file.Open()
	if err != nil {
		return "", "", nil, err
	}
	defer src.Close()

	content, err = io.ReadAll(src)
	if err != nil {
		return "", "", nil, err
	}

	head := content
	if len(content) > 8192 {
		head = content[:8192]
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
		return "", "", nil, errors.ErrUnsupported
	}

	if isAvatar {
		config, _, decodeErr := image.DecodeConfig(bytes.NewReader(content))
		if decodeErr != nil {
			return "", "", nil, decodeErr
		}
		if config.Height > maxImageHeight || config.Width > maxImageWidth {
			return "", "", nil, ErrInvalidDimensions
		}
	}

	switch {
	case isAvatar:
		subDir = "avatars"
	case isProductPicture:
		subDir = "products_pictures"
	case isProductVideo:
		subDir = "product_videos"
	case isResume:
		subDir = "resumes"
	}

	return fileName, subDir, content, nil
}
