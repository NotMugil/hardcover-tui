package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
	"github.com/NotMugil/hardcover-tui/internal/keystore"
)

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
		if len(args) < 3 {
			fmt.Println("Usage: hardcover-tui auth login <Bearer TOKEN>")
			os.Exit(1)
		}
		token := strings.TrimSpace(args[2])
		if !strings.HasPrefix(token, "Bearer ") {
			token = "Bearer " + token
		}
		if err := keystore.Save(token); err != nil {
			fmt.Printf("Error saving API token: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✔ API token saved to keyring successfully.")

	case "logout":
		if err := keystore.Delete(); err != nil {
			fmt.Printf("Error deleting API token: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✔ Logged out. API token removed from keyring.")

	case "status":
		token, err := keystore.Load()
		if err != nil || token == "" {
			fmt.Println("Status: Not authenticated (no token in keyring). Run 'hardcover-tui auth login <token>'")
			return true
		}
		fmt.Println("Status: Token found in keyring.")
		fmt.Println("Verifying token against Hardcover API...")
		c := api.NewClient(token)
		user, err := queries.GetMe(context.Background(), c)
		if err != nil {
			fmt.Printf("⚠ Token verification failed: %v\n", err)
		} else {
			fmt.Printf("✔ Authenticated as @%s (ID: %d)\n", user.Username, user.ID)
		}

	default:
		fmt.Println("Auth Subcommands:")
		fmt.Println("  hardcover-tui auth login <token>  Save API token")
		fmt.Println("  hardcover-tui auth logout         Remove saved token")
		fmt.Println("  hardcover-tui auth status         Check current auth status")
	}

	return true
}
