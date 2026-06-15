package fonts

import (
	"fmt"
	"os"
	"path/filepath"
)

type Service struct {
	Regular []byte
	Bold    []byte
	Italic  []byte
}

func NewService(assetsDir string) (*Service, error) {
	dir := filepath.Join(assetsDir, "fonts", "dejavu-sans")

	regular, err := os.ReadFile(filepath.Join(dir, "DejaVuSans.ttf"))
	if err != nil {
		return nil, fmt.Errorf("read regular font: %w", err)
	}
	bold, err := os.ReadFile(filepath.Join(dir, "DejaVuSans-Bold.ttf"))
	if err != nil {
		return nil, fmt.Errorf("read bold font: %w", err)
	}
	italic, err := os.ReadFile(filepath.Join(dir, "DejaVuSans-Oblique.ttf"))
	if err != nil {
		return nil, fmt.Errorf("read italic font: %w", err)
	}

	return &Service{Regular: regular, Bold: bold, Italic: italic}, nil
}
