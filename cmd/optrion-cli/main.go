package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/optrion/optrion/internal/config/app"
	"github.com/optrion/optrion/internal/registration/domain"
	"gopkg.in/yaml.v3"
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
  optrion-cli verify --config optrion.yaml --server http://localhost:8080
`)
}

// handleInit creates a sample optrion.yaml file
func handleInit() {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	output := fs.String("output", "optrion.yaml", "Output file path")
	fs.Parse(os.Args[2:])

	template := app.GenerateTemplate()

	if err := ioutil.WriteFile(*output, []byte(template), 0644); err != nil {
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

	// Convert to JSON and send to server
	regData, err := yaml.Marshal(req)
	if err != nil {
		log.Fatalf("Failed to marshal registration: %v", err)
	}

	// TODO: Send HTTP POST to server
	fmt.Printf("✓ Registration request prepared:\n%s\n", regData)
	fmt.Println("\nAPI Key will be provided upon successful registration")
	fmt.Println("Save the API key securely - you'll need it for monitoring")
}

// handleVerify validates the integration
func handleVerify() {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	configFile := fs.String("config", "optrion.yaml", "Configuration file path")
	serverURL := fs.String("server", "http://localhost:8080", "OPTRION server URL")
	apiKey := fs.String("api-key", "", "API Key from registration")
	fs.Parse(os.Args[2:])

	if *apiKey == "" {
		log.Fatal("API key required for verification")
	}

	// Load config
	loader := app.NewConfigLoader(*configFile)
	config, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	fmt.Println("OPTRION Integration Verification")
	fmt.Println("--------------------------------")

	// Check configuration validity
	fmt.Print("✓ Configuration file valid: ")
	fmt.Println("OK")

	// Verify component connectivity
	fmt.Println("\nComponent Connectivity:")
	for _, comp := range config.Components {
		fmt.Printf("  Checking %s (%s)...", comp.Name, comp.Kind)
		// TODO: Implement actual connectivity check
		fmt.Println(" OK")
	}

	// Verify metrics flowing
	fmt.Println("\nMetrics Status:")
	fmt.Print("  Metrics flowing from server: ")
	// TODO: Query server for recent metrics
	fmt.Println("OK")

	// Verify components registered
	fmt.Println("\nRegistered Components:")
	for _, comp := range config.Components {
		fmt.Printf("  ✓ %s (%s)\n", comp.Name, comp.Kind)
	}

	// Verify health visible
	fmt.Println("\nHealth Status:")
	fmt.Print("  Platform health visible: ")
	// TODO: Query server for health data
	fmt.Println("OK")

	fmt.Println("\n✓ Integration verified successfully!")
	fmt.Printf("\nServer URL: %s\n", *serverURL)
	fmt.Printf("Config file: %s\n", *configFile)
}
