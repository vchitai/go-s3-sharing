package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	baseURL := "http://localhost:8080"

	// Example 1: Create a share
	fmt.Println("=== Creating a Share ===")
	createShareReq := map[string]interface{}{
		"s3_path":    "images/example.jpg",
		"secret":     "my-secret-key",
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	reqBody, err := json.Marshal(createShareReq)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(baseURL+"/api/shares", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Failed to create share: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to create share: %s", string(body))
	}

	var createResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	fmt.Printf("Share created successfully!\n")
	fmt.Printf("URL: %s\n", createResp["url"])
	fmt.Printf("Expires at: %s\n", createResp["expires_at"])

	// Example 2: Access the shared file
	fmt.Println("\n=== Accessing Shared File ===")
	shareURL := createResp["url"].(string)

	// Extract the path from the URL for the GET request
	// Assuming the URL format is: http://localhost:8080/yy/mm/dd/secret/path
	// We need to extract the path part after the base URL
	path := shareURL[len(baseURL):]

	resp, err = http.Get(baseURL + path)
	if err != nil {
		log.Fatalf("Failed to access shared file: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %s\n", resp.Status)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Content-Length: %s\n", resp.Header.Get("Content-Length"))

	// Example 3: Health check
	fmt.Println("\n=== Health Check ===")
	resp, err = http.Get(baseURL + "/health")
	if err != nil {
		log.Fatalf("Failed to check health: %v", err)
	}
	defer resp.Body.Close()

	var healthResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		log.Fatalf("Failed to decode health response: %v", err)
	}

	fmt.Printf("Health status: %s\n", healthResp["status"])

	// Example 4: Readiness check
	fmt.Println("\n=== Readiness Check ===")
	resp, err = http.Get(baseURL + "/ready")
	if err != nil {
		log.Fatalf("Failed to check readiness: %v", err)
	}
	defer resp.Body.Close()

	var readyResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&readyResp); err != nil {
		log.Fatalf("Failed to decode readiness response: %v", err)
	}

	fmt.Printf("Readiness status: %s\n", readyResp["status"])
}
