package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
	"github.com/NotMugil/hardcover-tui/internal/app"
	"github.com/NotMugil/hardcover-tui/internal/auth"
	"github.com/NotMugil/hardcover-tui/internal/keystore"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("hardcover-tui %s\n", version)
		os.Exit(0)
	}

	if len(os.Args) > 1 && handleAuthCommand(os.Args[1:]) {
		os.Exit(0)
	}

	zone.NewGlobal()
	p := tea.NewProgram(app.New(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleAuthCommand(args []string) bool {
	if len(args) == 0 || args[0] != "auth" {
		return false
	}

	sub := ""
	if len(args) > 1 {
		sub = args[1]
	}

	switch sub {
	case "login":
		loginArgs := args[2:]
		handleAuthLogin(loginArgs)

	case "logout":
		if err := keystore.Delete(); err != nil {
			fmt.Printf("Error deleting API token: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✔ Logged out. API token removed from keyring.")

	case "status":
		token, err := keystore.Load()
		if err != nil || token == "" {
			fmt.Println("Status: Not authenticated (no token in keyring). Run 'hardcover-tui auth login'")
			return true
		}
		fmt.Println("Status: Token found in keyring.")
		fmt.Println("Verifying token against Hardcover API...")
		c := api.NewClient(token)
		user, err := queries.GetMe(context.Background(), c)
		if err != nil {
			var scopeErr *api.InsufficientScopeError
			switch {
			case errors.Is(err, api.ErrTokenRevoked):
				fmt.Println("⚠ Token verification failed: Token is invalid, expired, or has been revoked (HTTP 401).")
				fmt.Println("  Run 'hardcover-tui auth logout' and 'hardcover-tui auth login' to re-authenticate.")
			case errors.As(err, &scopeErr):
				fmt.Printf("⚠ Token verification warning: Token lacks required permissions (%s).\n", scopeErr.Scope)
				fmt.Println("  Ensure your token has 'all' (Full Access) scope.")
			case errors.Is(err, api.ErrDailyLimitExceeded):
				fmt.Println("⚠ Daily API request limit reached (5,000/50,000 requests). Resets at midnight UTC.")
			default:
				fmt.Printf("⚠ Token verification failed: %v\n", err)
			}
		} else {
			fmt.Printf("✔ Authenticated as @%s (ID: %d)\n", user.Username, user.ID)
		}

	default:
		fmt.Println("Auth Subcommands:")
		fmt.Println("  hardcover-tui auth login                  Interactive browser login (Device Authorization)")
		fmt.Println("  hardcover-tui auth login --token <token>  Save Personal Access Token manually")
		fmt.Println("  hardcover-tui auth logout                 Remove saved token")
		fmt.Println("  hardcover-tui auth status                 Check current auth status")
	}

	return true
}

func handleAuthLogin(args []string) {
	// Check if --token, -t, or --pat flag was passed
	isManualToken := false
	var manualToken string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--token" || arg == "-t" || arg == "--pat" || arg == "--manual" {
			isManualToken = true
			if i+1 < len(args) {
				manualToken = strings.TrimSpace(strings.Join(args[i+1:], " "))
			}
			break
		}
	}

	// Also support legacy: hardcover-tui auth login <TOKEN> (if first arg does not start with -)
	if !isManualToken && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		isManualToken = true
		manualToken = strings.TrimSpace(strings.Join(args, " "))
	}

	if isManualToken {
		if manualToken == "" {
			fmt.Println("Personal Access Token (PAT) Manual Login")
			fmt.Println("----------------------------------------")
			fmt.Println("Please generate an API key with 'all' (Full Access) scope at:")
			fmt.Println("  https://hardcover.app/account/api/keys/new?scope=all")
			fmt.Println()
			fmt.Print("Enter your API Token: ")
			var input string
			_, _ = fmt.Scanln(&input)
			manualToken = strings.TrimSpace(input)
		}

		token := keystore.FormatToken(manualToken)
		if token == "" {
			fmt.Println("Error: Invalid API token provided.")
			os.Exit(1)
		}
		if err := keystore.Save(token); err != nil {
			fmt.Printf("Error saving API token: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✔ API token saved to keyring successfully.")
		verifySavedToken(token)
		return
	}

	// Device Authorization Flow
	ctx := context.Background()
	clientID := auth.GetClientID()

	fmt.Println("Initiating login with Hardcover...")
	authResp, err := auth.RequestDeviceCode(ctx, clientID)
	if err != nil {
		fmt.Printf("Error requesting device authorization: %v\n", err)
		fmt.Println("\nFallback: You can log in manually using a Personal Access Token:")
		fmt.Println("  hardcover-tui auth login --token <YOUR_TOKEN>")
		fmt.Println("Generate one with 'all' scope at: https://hardcover.app/account/api/keys/new?scope=all")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("======================================================")
	fmt.Println("    Connect hardcover-tui to your Hardcover Account   ")
	fmt.Println("======================================================")
	fmt.Println()
	fmt.Printf("  Verification Code:  \033[1;36m%s\033[0m\n", authResp.UserCode)
	fmt.Println()
	fmt.Printf("  Opening browser to: \033[4m%s\033[0m\n", authResp.VerificationURIComplete)
	fmt.Println("  (If the browser does not open, visit the URL above manually)")
	fmt.Println()
	fmt.Println("Waiting for authorization in browser...")

	// Open browser automatically
	_ = auth.OpenBrowser(authResp.VerificationURIComplete)

	interval := time.Duration(authResp.Interval) * time.Second
	expiresAt := time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIdx := 0

	token, err := auth.PollToken(ctx, clientID, authResp.DeviceCode, interval, expiresAt, func() {
		fmt.Printf("\r%s Waiting for approval...", spinnerChars[spinnerIdx%len(spinnerChars)])
		spinnerIdx++
	})
	fmt.Print("\r\033[K") // clear spinner line

	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		os.Exit(1)
	}

	formattedToken := keystore.FormatToken(token)
	if err := keystore.Save(formattedToken); err != nil {
		fmt.Printf("Error saving token to keyring: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✔ Authenticated successfully! Token saved to keyring.")
	verifySavedToken(formattedToken)
}

func verifySavedToken(token string) {
	fmt.Println("Verifying token against Hardcover API...")
	c := api.NewClient(token)
	user, err := queries.GetMe(context.Background(), c)
	if err != nil {
		var scopeErr *api.InsufficientScopeError
		switch {
		case errors.Is(err, api.ErrTokenRevoked):
			fmt.Println("⚠ Warning: Token is invalid, expired, or has been revoked (HTTP 401).")
		case errors.As(err, &scopeErr):
			fmt.Printf("⚠ Warning: Token lacks required scope '%s'. Please generate a key with 'all' scope.\n", scopeErr.Scope)
		case errors.Is(err, api.ErrDailyLimitExceeded):
			fmt.Println("⚠ Note: Daily API request limit reached. Resets at midnight UTC.")
		default:
			fmt.Printf("⚠ Note: Token verification warning: %v\n", err)
		}
	} else {
		fmt.Printf("✔ Logged in as @%s (ID: %d)\n", user.Username, user.ID)
	}
}
