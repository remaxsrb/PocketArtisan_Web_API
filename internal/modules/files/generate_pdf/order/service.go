package order

import (
	"PocketArtisan/internal/modules/files/storage"
	"bytes"
	"fmt"
	"time"
)

type Service struct {
	storage storage.Storage
}

func NewService(s storage.Storage) *Service {
	return &Service{storage: s}
}

// Generate builds an order-confirmation PDF and saves it via the storage layer.
// It returns the public URL of the saved file.
func (s *Service) Generate(data OrderData) (string, error) {
	f, err := buildPDF(data)
	if err != nil {
		return "", fmt.Errorf("build pdf: %w", err)
	}

	var buf bytes.Buffer
	if err := f.Output(&buf); err != nil {
		return "", fmt.Errorf("render pdf: %w", err)
	}

	filename := fmt.Sprintf("order_%d_%d.pdf", data.OrderID, time.Now().UnixMilli())
	url, err := s.storage.SaveRawFile(buf.Bytes(), filename, "orders")
	if err != nil {
		return "", fmt.Errorf("save pdf: %w", err)
	}
	return url, nil
}
