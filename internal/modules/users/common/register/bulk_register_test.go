package register

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
	apiURL     = "http://localhost:8080/users/register"
	dataFile   = "userdata.json"
	maxWorkers = 5 // Limit concurrent requests to avoid overwhelming the server
)

func loadUsers(filename string) ([]RegisterRequest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var new_users []RegisterRequest
	if err := json.Unmarshal(data, &new_users); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return new_users, nil
}

func registerUser(t *testing.T, new_user RegisterRequest, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()

	// Acquire semaphore slot
	sem <- struct{}{}
	defer func() { <-sem }() // Release slot when done

	payload, err := json.Marshal(new_user)
	if err != nil {
		t.Errorf("Failed to marshal user %s: %v", new_user.Username, err)
		return
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Failed to create request for %s: %v", new_user.Username, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Network error for %s: %v", new_user.Username, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Logf("Success: %s (Status: %d)", new_user.Username, resp.StatusCode)
	} else {
		t.Errorf("Failed: %s (Status: %d) - %s", new_user.Username, resp.StatusCode, string(body))
	}
}

func TestBulkUserRegistration(t *testing.T) {
	users, err := loadUsers(dataFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if len(users) == 0 {
		t.Fatal("No users found in data file")
	}

	t.Logf("Starting bulk registration of %d users...", len(users))

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxWorkers) // Semaphore for concurrency control

	for _, user := range users {
		wg.Add(1)
		go registerUser(t, user, &wg, sem)

		// Optional: Small delay between launching goroutines to smooth out burst
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()
	t.Log("Bulk registration test completed.")
}
