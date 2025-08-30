package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/wukong0111/go-banks/internal/auth"
	"github.com/wukong0111/go-banks/internal/config"
	"github.com/wukong0111/go-banks/internal/logger"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist, just warn
		fmt.Fprintf(os.Stderr, "Warning: Could not load .env file: %v\n", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Define command line flags
	var (
		apiKeyFlag      = flag.String("apikey", "", "API key for authentication (defaults to API_KEY env var)")
		permissionsFlag = flag.String("permissions", "banks:read", "Comma-separated list of permissions (e.g., banks:read,banks:write)")
		expiryFlag      = flag.String("expiry", "", "Token expiry duration (e.g., 24h, 1h, 30m) - defaults to JWT_EXPIRY env var")
		helpFlag        = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	// Show help if requested
	if *helpFlag {
		showHelp()
		return
	}

	// Validate API key
	apiKey := *apiKeyFlag
	if apiKey == "" {
		apiKey = cfg.APIKey
	}
	if apiKey != cfg.APIKey {
		log.Fatalf("Invalid API key. Use -apikey flag or set API_KEY environment variable")
	}

	// Parse permissions
	permissionsList := parsePermissions(*permissionsFlag)
	if len(permissionsList) == 0 {
		log.Fatalf("At least one permission is required")
	}

	// Validate permissions
	if !validatePermissions(permissionsList) {
		log.Fatalf("Invalid permissions. Allowed: banks:read, banks:write")
	}

	// Parse expiry duration
	var expiry time.Duration
	if *expiryFlag != "" {
		expiry, err = time.ParseDuration(*expiryFlag)
		if err != nil {
			log.Fatalf("Invalid expiry duration: %v", err)
		}
	} else {
		expiry, err = time.ParseDuration(cfg.JWT.Expiry)
		if err != nil {
			log.Fatalf("Invalid JWT_EXPIRY config: %v", err)
		}
	}

	// Create JWT service
	tokenLogger := logger.NewMultiLogger(slog.NewTextHandler(io.Discard, nil))
	jwtService := auth.NewJWTService(cfg.JWT.Secret, expiry, tokenLogger)

	// Generate token
	token, err := jwtService.GenerateToken(permissionsList)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(expiry)

	// Output success
	fmt.Println("🔑 Token generated successfully!")
	fmt.Println()
	fmt.Printf("Token: %s\n", token)
	fmt.Printf("Expires: %s\n", expiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Permissions: %v\n", permissionsList)
	fmt.Printf("Duration: %s\n", expiry)
	fmt.Println()
	fmt.Println("Use this token in the Authorization header:")
	fmt.Printf("Authorization: Bearer %s\n", token)
}

func parsePermissions(permissionsStr string) []string {
	if permissionsStr == "" {
		return []string{}
	}

	permissions := strings.Split(permissionsStr, ",")
	var result []string

	for _, perm := range permissions {
		perm = strings.TrimSpace(perm)
		if perm != "" {
			result = append(result, perm)
		}
	}

	return result
}

func validatePermissions(permissions []string) bool {
	allowedPermissions := map[string]bool{
		"banks:read":  true,
		"banks:write": true,
	}

	for _, perm := range permissions {
		if !allowedPermissions[perm] {
			return false
		}
	}

	return true
}

func showHelp() {
	fmt.Println("JWT Token Generator for Go Banks API")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  go run cmd/token/main.go [options]\n")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -apikey string")
	fmt.Println("        API key for authentication (defaults to API_KEY env var)")
	fmt.Println("  -permissions string")
	fmt.Println("        Comma-separated list of permissions (default: banks:read)")
	fmt.Println("        Available permissions: banks:read, banks:write")
	fmt.Println("  -expiry string")
	fmt.Println("        Token expiry duration (defaults to JWT_EXPIRY env var)")
	fmt.Println("        Examples: 24h, 1h, 30m, 1h30m")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate token with read permissions (default)")
	fmt.Println("  go run cmd/token/main.go")
	fmt.Println()
	fmt.Println("  # Generate token with read and write permissions")
	fmt.Println("  go run cmd/token/main.go -permissions banks:read,banks:write")
	fmt.Println()
	fmt.Println("  # Generate token with custom expiry")
	fmt.Println("  go run cmd/token/main.go -expiry 1h")
	fmt.Println()
	fmt.Println("  # Generate token with custom API key")
	fmt.Println("  go run cmd/token/main.go -apikey your-api-key -permissions banks:write")
}
