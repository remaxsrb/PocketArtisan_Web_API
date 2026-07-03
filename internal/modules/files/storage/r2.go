package storage

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/h2non/filetype"
)

// R2Storage stores files in a Cloudflare R2 bucket. R2 exposes an
// S3-compatible API, so the AWS SDK v2 S3 client is used under the hood.
type R2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// NewR2Storage builds an R2-backed Storage using the account credentials and
// S3-compatible endpoint. publicURL is the base URL used to build public file
// links (e.g. a custom domain or the bucket dev URL); when empty it falls back
// to "<endpoint>/<bucket>".
func NewR2Storage(accountID, accessKeyID, secretAccessKey, bucket, endpoint, publicURL string) (*R2Storage, error) {
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	if publicURL == "" {
		publicURL = strings.TrimRight(endpoint, "/") + "/" + bucket
	}

	return &R2Storage{
		client:    client,
		bucket:    bucket,
		publicURL: strings.TrimRight(publicURL, "/"),
	}, nil
}

func (r *R2Storage) put(ctx context.Context, key string, content []byte) error {
	contentType := "application/octet-stream"
	if kind, _ := filetype.Match(content); kind != filetype.Unknown {
		contentType = kind.MIME.Value
	}

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	return err
}

func (r *R2Storage) SaveFile(file *multipart.FileHeader, purpose string) (string, error) {
	fileName, subDir, content, err := resolveUpload(file, purpose)
	if err != nil {
		return "", err
	}

	key := subDir + "/" + fileName

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.put(ctx, key, content); err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	return r.publicURL + "/" + key, nil
}

func (r *R2Storage) SaveRawFile(data []byte, filename, subDir string) (string, error) {
	key := filename
	if subDir != "" {
		key = subDir + "/" + filename
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.put(ctx, key, data); err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	return r.publicURL + "/" + key, nil
}

func (r *R2Storage) DeleteFile(fileName string) error {
	key := r.normalizeKey(fileName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}
	return nil
}

func (r *R2Storage) GetFileURL(fileName string) string {
	return r.publicURL + "/" + r.normalizeKey(fileName)
}

// normalizeKey strips a leading slash and the public URL prefix so callers may
// pass either a bare object key or a full public URL.
func (r *R2Storage) normalizeKey(fileName string) string {
	key := strings.TrimPrefix(fileName, r.publicURL+"/")
	return strings.TrimPrefix(key, "/")
}

// ensure R2Storage satisfies the Storage interface
var _ Storage = (*R2Storage)(nil)
