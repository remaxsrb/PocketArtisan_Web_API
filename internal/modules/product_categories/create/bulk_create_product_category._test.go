package create

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	productCategoryCreateAPIURL = "http://localhost:8080/product-categories/create"
	categoryDataFile            = "categories_data.json"
	categoryMaxWorkers          = 5
)

func loadProductCategories(filename string) ([]NewProductCategoryRequest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var categories []NewProductCategoryRequest
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return categories, nil
}

func createProductCategory(t *testing.T, category NewProductCategoryRequest, adminToken string, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()

	sem <- struct{}{}
	defer func() { <-sem }()

	payload, err := json.Marshal(category)
	if err != nil {
		t.Errorf("Failed to marshal product category %s: %v", category.Name, err)
		return
	}

	req, err := http.NewRequest("POST", productCategoryCreateAPIURL, bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request for %s: %v", category.Name, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Network error for %s: %v", category.Name, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Logf("Success: %s (Status: %d)", category.Name, resp.StatusCode)
	} else {
		t.Errorf("Failed: %s (Status: %d) - %s", category.Name, resp.StatusCode, string(body))
	}
}

func TestBulkProductCategoryCreate(t *testing.T) {
	adminToken := os.Getenv("ADMIN_BEARER_TOKEN")
	if adminToken == "" {
		t.Skip("Skipping bulk product category create test: ADMIN_BEARER_TOKEN is not set")
	}

	categories, err := loadProductCategories(categoryDataFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if len(categories) == 0 {
		t.Fatal("No product categories found in data file")
	}

	t.Logf("Starting bulk create of %d product categories...", len(categories))

	var wg sync.WaitGroup
	sem := make(chan struct{}, categoryMaxWorkers)

	for _, category := range categories {
		wg.Add(1)
		go createProductCategory(t, category, adminToken, &wg, sem)
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()
	t.Log("Bulk product category create test completed.")
}
