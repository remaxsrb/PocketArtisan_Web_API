package approve

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
	CApproveAPIURL = "http://localhost:8080/api/admin/craftsman-applications/approve"
	caDataFile     = "approve_data.json"
	caMaxWorkers   = 5
)

func loadIDs(filename string) ([]Request, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var requests []Request
	if err := json.Unmarshal(data, &requests); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return requests, nil

}

func approveApplication(t *testing.T, approval Request, adminToken string, wg *sync.WaitGroup, sem chan struct{}) {

	defer wg.Done()
	sem <- struct{}{}
	defer func() { <-sem }()

	payload, err := json.Marshal(approval)
	if err != nil {
		t.Errorf("Failed to marshal id %d: %v", approval.ApplicationID, err)
		return
	}

	req, err := http.NewRequest("PATCH", CApproveAPIURL, bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request for %d: %v", approval.ApplicationID, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Network error for %d: %v", approval.ApplicationID, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Logf("Success: %d (Status: %d)", approval.ApplicationID, resp.StatusCode)
	} else {
		t.Errorf("Failed: %d (Status: %d) - %s", approval.ApplicationID, resp.StatusCode, string(body))
	}

}

func TestBulkApproveCraftsmanApplication(t *testing.T) {
	adminToken := os.Getenv("ADMIN_BEARER_TOKEN")
	if adminToken == "" {
		t.Skip("Skipping bulk approve test: ADMIN_BEARER_TOKEN is not set")
	}

	requests, err := loadIDs(caDataFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if len(requests) == 0 {
		t.Fatal("No ids found in data file")
	}

	t.Logf("Starting bulk approval of %d applications...", len(requests))

	var wg sync.WaitGroup
	sem := make(chan struct{}, caMaxWorkers)

	for _, request := range requests {
		wg.Add(1)
		go approveApplication(t, request, adminToken, &wg, sem)
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()
	t.Log("Bulk craft create test completed.")
}
