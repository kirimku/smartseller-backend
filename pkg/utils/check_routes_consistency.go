package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"

	"gopkg.in/yaml.v3"
)

type OpenAPI struct {
	Paths map[string]interface{} `yaml:"paths"`
}

func main() {
	// Read routes from main.go
	mainFile, err := ioutil.ReadFile("cmd/main.go")
	if err != nil {
		log.Fatalf("Failed to read main.go: %v", err)
	}

	// Extract routes from main.go
	mainRoutes := extractRoutes(string(mainFile))

	// Read routes from openapi.yaml
	openapiFile, err := ioutil.ReadFile("openapi.yaml")
	if err != nil {
		log.Fatalf("Failed to read openapi.yaml: %v", err)
	}

	var openapi OpenAPI
	if err := yaml.Unmarshal(openapiFile, &openapi); err != nil {
		log.Fatalf("Failed to parse openapi.yaml: %v", err)
	}

	openapiRoutes := extractOpenAPIRoutes(openapi)

	// Compare routes
	checkConsistency(mainRoutes, openapiRoutes)
}

func extractRoutes(fileContent string) []string {
	routeRegex := regexp.MustCompile(`http\.HandleFunc\("([^"]+)"`)
	matches := routeRegex.FindAllStringSubmatch(fileContent, -1)

	var routes []string
	for _, match := range matches {
		if len(match) > 1 {
			routes = append(routes, match[1])
		}
	}
	return routes
}

func extractOpenAPIRoutes(openapi OpenAPI) []string {
	var routes []string
	for route := range openapi.Paths {
		routes = append(routes, route)
	}
	return routes
}

func checkConsistency(mainRoutes, openapiRoutes []string) {
	mainRouteSet := make(map[string]bool)
	for _, route := range mainRoutes {
		mainRouteSet[route] = true
	}

	openapiRouteSet := make(map[string]bool)
	for _, route := range openapiRoutes {
		openapiRouteSet[route] = true
	}

	// Check for missing routes in OpenAPI
	fmt.Println("Checking for missing routes in OpenAPI...")
	for _, route := range mainRoutes {
		if !openapiRouteSet[route] {
			fmt.Printf("Route missing in OpenAPI: %s\n", route)
		}
	}

	// Check for extra routes in OpenAPI
	fmt.Println("\nChecking for extra routes in OpenAPI...")
	for _, route := range openapiRoutes {
		if !mainRouteSet[route] {
			fmt.Printf("Extra route in OpenAPI: %s\n", route)
		}
	}
}
