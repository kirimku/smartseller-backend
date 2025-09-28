package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	EmailOrPhone string `json:"email_or_phone"`
	Password     string `json:"password"`
}

// LoginResponse represents the response from the login endpoint
type LoginResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenExpiry  string `json:"token_expiry"`
		User         struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	} `json:"data"`
	Meta struct {
		HTTPStatus int `json:"http_status"`
	} `json:"meta"`
}

func main() {
	// Command line flags
	var (
		email    = flag.String("email", "", "Email or phone number for login")
		password = flag.String("password", "", "Password for login")
		baseURL  = flag.String("url", "http://localhost:8080", "Base URL of the API server")
		output   = flag.String("output", "json", "Output format: json, token, or env")
		help     = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Use environment variables if flags are not provided
	if *email == "" {
		*email = os.Getenv("LOGIN_EMAIL")
	}
	if *password == "" {
		*password = os.Getenv("LOGIN_PASSWORD")
	}

	// Validate required parameters
	if *email == "" || *password == "" {
		fmt.Fprintf(os.Stderr, "Error: Email and password are required\n")
		fmt.Fprintf(os.Stderr, "Use flags: -email=user@example.com -password=yourpassword\n")
		fmt.Fprintf(os.Stderr, "Or set environment variables: LOGIN_EMAIL and LOGIN_PASSWORD\n")
		fmt.Fprintf(os.Stderr, "Use -help for more information\n")
		os.Exit(1)
	}

	// Perform login
	token, response, err := performLogin(*email, *password, *baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		os.Exit(1)
	}

	// Output based on format
	switch *output {
	case "token":
		fmt.Print(token)
	case "env":
		fmt.Printf("export ACCESS_TOKEN=%s\n", token)
		fmt.Printf("export REFRESH_TOKEN=%s\n", response.Data.RefreshToken)
		fmt.Printf("export USER_ID=%s\n", response.Data.User.ID)
		fmt.Printf("export TOKEN_EXPIRY=%s\n", response.Data.TokenExpiry)
	case "json":
		fallthrough
	default:
		jsonOutput, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonOutput))
	}
}

func performLogin(email, password, baseURL string) (string, *LoginResponse, error) {
	// Create login request
	loginReq := LoginRequest{
		EmailOrPhone: email,
		Password:     password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	loginURL := fmt.Sprintf("%s/api/v1/auth/login", baseURL)
	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if login was successful
	if !loginResp.Success {
		return "", nil, fmt.Errorf("login failed: %s", loginResp.Message)
	}

	return loginResp.Data.AccessToken, &loginResp, nil
}

func showHelp() {
	fmt.Println("Platform Login Script")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("This script authenticates with the SmartSeller platform and obtains access tokens for testing.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run scripts/login.go [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -email string     Email or phone number for login")
	fmt.Println("  -password string  Password for login")
	fmt.Println("  -url string       Base URL of the API server (default: http://localhost:8080)")
	fmt.Println("  -output string    Output format: json, token, or env (default: json)")
	fmt.Println("  -help            Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  LOGIN_EMAIL      Default email if -email flag is not provided")
	fmt.Println("  LOGIN_PASSWORD   Default password if -password flag is not provided")
	fmt.Println()
	fmt.Println("Output Formats:")
	fmt.Println("  json    Full JSON response (default)")
	fmt.Println("  token   Only the access token")
	fmt.Println("  env     Environment variable export statements")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Login and get full JSON response")
	fmt.Println("  go run scripts/login.go -email=user@example.com -password=mypassword")
	fmt.Println()
	fmt.Println("  # Get only the access token")
	fmt.Println("  go run scripts/login.go -email=user@example.com -password=mypassword -output=token")
	fmt.Println()
	fmt.Println("  # Get environment variables to source")
	fmt.Println("  go run scripts/login.go -email=user@example.com -password=mypassword -output=env")
	fmt.Println()
	fmt.Println("  # Use environment variables for credentials")
	fmt.Println("  export LOGIN_EMAIL=user@example.com")
	fmt.Println("  export LOGIN_PASSWORD=mypassword")
	fmt.Println("  go run scripts/login.go")
	fmt.Println()
	fmt.Println("  # Source environment variables directly")
	fmt.Println("  eval $(go run scripts/login.go -email=user@example.com -password=mypassword -output=env)")
}