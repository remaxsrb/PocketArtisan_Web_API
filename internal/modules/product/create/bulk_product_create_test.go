package create

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	fileUploadAPIURL    = "http://localhost:8080/api/files/upload"
	productCreateAPIURL = "http://localhost:8080/api/products/create"
)

type productEntry struct {
	Username    string  `json:"username"`
	Craft       string  `json:"craft"`
	Category    string  `json:"category"`
	AssetFolder string  `json:"assetFolder"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func uploadFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", filePath, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return "", err
	}
	if err := w.WriteField("purpose", "product_image"); err != nil {
		return "", err
	}
	w.Close()

	req, err := http.NewRequest("POST", fileUploadAPIURL, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload failed (%d): %s", resp.StatusCode, body)
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w — body: %s", err, body)
	}
	if result.URL == "" {
		return "", fmt.Errorf("empty URL in response: %s", body)
	}
	return result.URL, nil
}

// TestBulkFileUpload uploads every file found in the assets directory.
// purpose is always product_photo.
func TestBulkFileUpload(t *testing.T) {
	assetsDir := os.Getenv("ASSETS_DIR")
	if assetsDir == "" {
		t.Skip("Skipping: ASSETS_DIR not set")
	}

	var filePaths []string
	folders, err := os.ReadDir(assetsDir)
	if err != nil {
		t.Fatalf("Failed to read assets dir: %v", err)
	}
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}
		subDir := filepath.Join(assetsDir, folder.Name())
		files, err := os.ReadDir(subDir)
		if err != nil {
			t.Fatalf("Failed to read subdir %s: %v", folder.Name(), err)
		}
		for _, file := range files {
			if !file.IsDir() {
				filePaths = append(filePaths, filepath.Join(subDir, file.Name()))
			}
		}
	}

	if len(filePaths) == 0 {
		t.Fatal("No files found in assets directory")
	}
	t.Logf("Uploading %d files...", len(filePaths))

	success := 0
	for _, p := range filePaths {
		url, err := uploadFile(p)
		if err != nil {
			t.Errorf("Failed %s: %v", p, err)
			continue
		}
		t.Logf("Uploaded %s → %s", filepath.Base(p), url)
		success++
	}

	t.Logf("Done: %d/%d files uploaded successfully.", success, len(filePaths))
}

// TestSingleProductCreate uploads ALL images in the product's asset folder
// sequentially, waits for every URL to come back, then creates the product
// with the full Images slice. Called once per product by the init script.
func TestSingleProductCreate(t *testing.T) {
	assetsDir := os.Getenv("ASSETS_DIR")
	if assetsDir == "" {
		t.Skip("Skipping: ASSETS_DIR not set")
	}
	bearerToken := os.Getenv("BEARER_TOKEN")
	if bearerToken == "" {
		t.Skip("Skipping: BEARER_TOKEN not set")
	}
	productJSON := os.Getenv("PRODUCT_JSON")
	if productJSON == "" {
		t.Skip("Skipping: PRODUCT_JSON not set")
	}

	var entry productEntry
	if err := json.Unmarshal([]byte(productJSON), &entry); err != nil {
		t.Fatalf("Failed to parse PRODUCT_JSON: %v", err)
	}

	// Collect all files in the asset folder for this product.
	folderPath := filepath.Join(assetsDir, entry.AssetFolder)
	dirEntries, err := os.ReadDir(folderPath)
	if err != nil {
		t.Fatalf("Failed to read folder %s: %v", folderPath, err)
	}

	var filePaths []string
	for _, f := range dirEntries {
		if !f.IsDir() {
			filePaths = append(filePaths, filepath.Join(folderPath, f.Name()))
		}
	}
	if len(filePaths) == 0 {
		t.Fatalf("No files found in folder %s", folderPath)
	}

	// Upload all images sequentially — product creation only runs once every
	// URL has been collected.
	t.Logf("Uploading %d image(s) for product %q...", len(filePaths), entry.Name)
	var imageURLs []string
	for _, p := range filePaths {
		url, err := uploadFile(p)
		if err != nil {
			t.Fatalf("Image upload failed for %s: %v", filepath.Base(p), err)
		}
		t.Logf("  Uploaded %s → %s", filepath.Base(p), url)
		imageURLs = append(imageURLs, url)
	}

	// All uploads done — now create the product.
	product := NewProductRequest{
		Name:        entry.Name,
		Description: entry.Description,
		Price:       entry.Price,
		Username:    entry.Username,
		Category:    entry.Category,
		Images:      imageURLs,
	}

	payload, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	req, err := http.NewRequest("POST", productCreateAPIURL, bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Request build failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Network error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Failed: %s (Status: %d) — %s", entry.Name, resp.StatusCode, string(body))
	}
	t.Logf("Created: %s (Status: %d)", entry.Name, resp.StatusCode)
}
