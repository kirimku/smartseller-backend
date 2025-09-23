package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a JWT token as argument")
		os.Exit(1)
	}

	token := os.Args[1]
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		fmt.Println("Invalid JWT token format")
		os.Exit(1)
	}

	// Ensure padding is correct for base64 decoding
	payload := parts[1]
	if len(payload)%4 != 0 {
		padding := 4 - (len(payload) % 4)
		payload = payload + strings.Repeat("=", padding)
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// Try raw URL encoding if standard URL encoding fails
		decoded, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			fmt.Printf("Error decoding payload: %v\n", err)
			os.Exit(1)
		}
	}

	// Pretty print the JSON
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(decoded, &prettyJSON); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		fmt.Println("Raw decoded payload:", string(decoded))
		os.Exit(1)
	}

	// Marshal with indentation
	prettyOutput, _ := json.MarshalIndent(prettyJSON, "", "  ")
	fmt.Println(string(prettyOutput))
}
