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
	craftCreateAPIURL = "http://localhost:8080/api/crafts/create"
	craftDataFile     = "raw_craft_data.json"
	craftMaxWorkers   = 5
)

func loadCrafts(filename string) ([]NewCraftRequest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var crafts []NewCraftRequest
	if err := json.Unmarshal(data, &crafts); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return crafts, nil
}

func createCraft(t *testing.T, craft NewCraftRequest, adminToken string, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()

	sem <- struct{}{}
	defer func() { <-sem }()

	payload, err := json.Marshal(craft)
	if err != nil {
		t.Errorf("Failed to marshal craft %s: %v", craft.Name, err)
		return
	}

	req, err := http.NewRequest("POST", craftCreateAPIURL, bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request for %s: %v", craft.Name, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Network error for %s: %v", craft.Name, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Logf("Success: %s (Status: %d)", craft.Name, resp.StatusCode)
	} else {
		t.Errorf("Failed: %s (Status: %d) - %s", craft.Name, resp.StatusCode, string(body))
	}
}

func TestBulkCraftCreate(t *testing.T) {
	adminToken := os.Getenv("ADMIN_BEARER_TOKEN")
	if adminToken == "" {
		t.Skip("Skipping bulk craft create test: ADMIN_BEARER_TOKEN is not set")
	}

	crafts, err := loadCrafts(craftDataFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if len(crafts) == 0 {
		t.Fatal("No crafts found in data file")
	}

	t.Logf("Starting bulk create of %d crafts...", len(crafts))

	var wg sync.WaitGroup
	sem := make(chan struct{}, craftMaxWorkers)

	for _, craft := range crafts {
		wg.Add(1)
		go createCraft(t, craft, adminToken, &wg, sem)
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()
	t.Log("Bulk craft create test completed.")
}
