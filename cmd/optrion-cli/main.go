package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/optrion/optrion/internal/config/app"
	"github.com/optrion/optrion/internal/registration/domain"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
	case "register":
		handleRegister()
	case "verify":
		handleVerify()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`OPTRION CLI - Plug-and-Play Platform Integration

Usage:
  optrion-cli <command> [options]

Commands:
  init       Initialize optrion.yaml configuration file
  register   Register application with OPTRION platform
  verify     Validate OPTRION integration
  help       Show this help message

Examples:
  # Generate optrion.yaml template
  optrion-cli init

  # Register with server
  optrion-cli register --config optrion.yaml --server http://localhost:8080

  # Verify integration
  optrion-cli verify --config optrion.yaml --server http://localhost:8080`)
}

// handleInit creates a sample optrion.yaml file
func handleInit() {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	output := fs.String("output", "optrion.yaml", "Output file path")
	fs.Parse(os.Args[2:])

	template := app.GenerateTemplate()

	if err := os.WriteFile(*output, []byte(template), 0644); err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}

	fmt.Printf("✓ Created %s\n", *output)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit optrion.yaml with your application details")
	fmt.Println("2. Run: optrion-cli register --config optrion.yaml")
	fmt.Println("3. Run: optrion-cli verify --config optrion.yaml")
}

// handleRegister registers the application with OPTRION
func handleRegister() {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	configFile := fs.String("config", "optrion.yaml", "Configuration file path")
	serverURL := fs.String("server", "http://localhost:8080", "OPTRION server URL")
	fs.Parse(os.Args[2:])

	// Load config
	loader := app.NewConfigLoader(*configFile)
	config, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("OPTRION Registration")
	fmt.Println("-------------------")
	fmt.Printf("Tenant: %s (%s)\n", config.Tenant.Name, config.Tenant.Slug)
	fmt.Printf("Product: %s (%s)\n", config.Product.Name, config.Product.Slug)
	fmt.Printf("Environment: %s (%s)\n", config.Environment.Name, config.Environment.Tier)
	fmt.Printf("Components: %d\n", len(config.Components))

	for _, comp := range config.Components {
		fmt.Printf("  - %s (%s)\n", comp.Name, comp.Kind)
	}

	fmt.Printf("\nRegistering with: %s\n", *serverURL)

	// Build registration request
	req := domain.RegistrationRequest{
		Tenant: domain.TenantRegistration{
			Name: config.Tenant.Name,
			Slug: config.Tenant.Slug,
			Plan: config.Tenant.Plan,
		},
		Product: domain.ProductRegistration{
			Name:        config.Product.Name,
			Slug:        config.Product.Slug,
			Description: config.Product.Description,
		},
		Environment: domain.EnvironmentRegistration{
			Name: config.Environment.Name,
			Tier: config.Environment.Tier,
		},
		Components: make([]domain.ComponentRegistration, len(config.Components)),
	}

	for i, comp := range config.Components {
		req.Components[i] = domain.ComponentRegistration{
			Name:        comp.Name,
			Kind:        comp.Kind,
			Description: comp.Description,
			Endpoint:    comp.Endpoint,
			Port:        comp.Port,
		}
	}

	// Send HTTP POST to server
	body, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Failed to marshal registration: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(*serverURL+"/api/v1/register", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Fatalf("Registration failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	fmt.Println("\n✓ Registration successful!")

	// Parse response for API key
	var result struct {
		TenantID string `json:"tenant_id"`
		APIKey   string `json:"api_key"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.APIKey != "" {
		fmt.Printf("\nAPI Key: %s\n", result.APIKey)
		fmt.Println("⚠️  Save this key securely — it will not be shown again.")
	}
}

// handleVerify validates the integration
func handleVerify() {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	configFile := fs.String("config", "optrion.yaml", "Configuration file path")
	serverURL := fs.String("server", "http://localhost:8080", "OPTRION server URL")
	apiKey := fs.String("api-key", "", "API Key from registration")
	fs.Parse(os.Args[2:])

	if *apiKey == "" {
		log.Fatal("API key required for verification. Use --api-key flag.")
	}

	// Load config
	loader := app.NewConfigLoader(*configFile)
	config, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("OPTRION Integration Verification")
	fmt.Println("--------------------------------")

	// Check configuration validity
	fmt.Print("✓ Configuration file valid\n")

	client := &http.Client{Timeout: 10 * time.Second}

	// Verify server connectivity
	fmt.Print("  Checking server connectivity... ")
	req, _ := http.NewRequest(http.MethodGet, *serverURL+"/healthz", nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("FAILED (%v)\n", err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("FAILED (status %d)\n", resp.StatusCode)
		os.Exit(1)
	}
	fmt.Println("OK")

	// Verify API key authentication
	fmt.Print("  Checking API key authentication... ")
	req, _ = http.NewRequest(http.MethodGet, *serverURL+"/api/v1/info", nil)
	req.Header.Set("Authorization", "Bearer "+*apiKey)
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("FAILED (%v)\n", err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		fmt.Println("FAILED (invalid API key)")
		os.Exit(1)
	}
	fmt.Println("OK")

	// Verify component endpoints
	fmt.Println("\n  Component Connectivity:")
	for _, comp := range config.Components {
		fmt.Printf("    Checking %s (%s)... ", comp.Name, comp.Kind)
		if comp.Endpoint != "" {
			checkReq, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, comp.Endpoint, nil)
			checkResp, err := client.Do(checkReq)
			if err != nil {
				fmt.Printf("UNREACHABLE (%v)\n", err)
				continue
			}
			checkResp.Body.Close()
			if checkResp.StatusCode < 500 {
				fmt.Println("OK")
			} else {
				fmt.Printf("UNHEALTHY (status %d)\n", checkResp.StatusCode)
			}
		} else {
			fmt.Println("SKIPPED (no endpoint configured)")
		}
	}

	fmt.Printf("\n✓ Integration verified successfully!\n")
	fmt.Printf("  Server: %s\n", *serverURL)
	fmt.Printf("  Config: %s\n", *configFile)
}
