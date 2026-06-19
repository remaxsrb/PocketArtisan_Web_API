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
	CACreateAPIURL = "http://localhost:8080/api/craftsman-applications/create"
	caDataFile     = "ca_requests.json"
	caMaxWorkers   = 5
)

func loadCraftsmanApplications(filename string) ([]CraftsmanApplicationRequest, error) {
	data, err := os.ReadFile(filename)

	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filename, err)
	}

	var cas []CraftsmanApplicationRequest

	if err := json.Unmarshal(data, &cas); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}
	return cas, nil
}

func createCraftsmanApplication(t *testing.T, ca CraftsmanApplicationRequest, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()

	sem <- struct{}{}
	defer func() { <-sem }()

	payload, err := json.Marshal(ca)
	if err != nil {
		t.Errorf("error marshalling craftsman-application request: %v", err)
		return
	}

	req, err := http.NewRequest("POST", CACreateAPIURL, bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create craftsman-application request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Failed to create craftsman-application request for %s: %v", ca.Email, err)
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Logf("Success: %s (Status: %d)", ca.Email, resp.StatusCode)
	} else {
		t.Errorf("Failed: %s (Status: %d) - %s", ca.Email, resp.StatusCode, string(body))
	}
}

func TestBulkCreateCraftsmanApplication(t *testing.T) {
	cas, err := loadCraftsmanApplications(caDataFile)
	if err != nil {
		t.Errorf("Setup failed: %v", err)
	}

	t.Logf("Starting bulk create of %d craftsman applications...", len(cas))

	var wg sync.WaitGroup
	sem := make(chan struct{}, caMaxWorkers)

	for _, ca := range cas {
		wg.Add(1)
		go createCraftsmanApplication(t, ca, &wg, sem)
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()
	t.Logf("Finished bulk create of %d craftsman applications", len(cas))

}
