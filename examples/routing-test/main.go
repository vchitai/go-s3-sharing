package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	baseURL := "http://localhost:8080"

	fmt.Println("=== Testing Go S3 Sharing Routing Fix ===")
	fmt.Println()

	// Test 1: Health endpoint should work
	fmt.Println("1. Testing /health endpoint...")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == 200 && bytes.Contains(body, []byte("healthy")) {
			fmt.Println("   ✅ Health endpoint working correctly")
		} else {
			fmt.Printf("   ❌ Unexpected response: %d - %s\n", resp.StatusCode, string(body))
		}
	}

	// Test 2: Ready endpoint should work
	fmt.Println("2. Testing /ready endpoint...")
	resp, err = http.Get(baseURL + "/ready")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == 200 && bytes.Contains(body, []byte("ready")) {
			fmt.Println("   ✅ Ready endpoint working correctly")
		} else {
			fmt.Printf("   ❌ Unexpected response: %d - %s\n", resp.StatusCode, string(body))
		}
	}

	// Test 3: API shares endpoint should work (if server is running with dependencies)
	fmt.Println("3. Testing /api/shares endpoint...")
	reqBody := map[string]interface{}{
		"s3_path": "images/test.jpg",
		"secret":  "test-secret",
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err = http.Post(baseURL+"/api/shares", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == 200 {
			fmt.Println("   ✅ API shares endpoint working correctly")
		} else if resp.StatusCode == 500 {
			fmt.Println("   ⚠️  API shares endpoint reachable but server needs Redis/S3 (expected)")
		} else {
			fmt.Printf("   ❌ Unexpected response: %d - %s\n", resp.StatusCode, string(body))
		}
	}

	// Test 4: Invalid paths should return 404
	fmt.Println("4. Testing invalid paths...")
	testPaths := []string{"/", "/invalid", "/api/invalid", "/some/random/path"}

	for _, path := range testPaths {
		resp, err := http.Get(baseURL + path)
		if err != nil {
			fmt.Printf("   ❌ Error for %s: %v\n", path, err)
		} else {
			resp.Body.Close()
			if resp.StatusCode == 404 {
				fmt.Printf("   ✅ %s correctly returns 404\n", path)
			} else {
				fmt.Printf("   ❌ %s returned %d (expected 404)\n", path, resp.StatusCode)
			}
		}
	}

	fmt.Println()
	fmt.Println("=== Routing Test Complete ===")
	fmt.Println("If you see ✅ for all tests, the routing fix is working correctly!")
	fmt.Println()
	fmt.Println("To run the server for testing:")
	fmt.Println("  export S3_BUCKET=your-bucket")
	fmt.Println("  export REDIS_ADDR=localhost:6379")
	fmt.Println("  ./server")
}
