package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [account-name] [path]",
	Short: "Switch git configuration to use a specific GitHub account",
	Long: `Switch git configuration to use a specific GitHub account.

Examples:
  krakn use personal              # Switch to personal account globally
  krakn use work ~/my-project     # Switch to work account for specific repository
  krakn use personal --global     # Explicitly set global configuration
  krakn use personal -g           # Same as --global (shorthand)

By default, switches globally unless a path is provided.
Use --global flag to explicitly set global configuration.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountName := args[0]
		var repoPath string
		
		// Check if --global flag is set
		globalFlag, _ := cmd.Flags().GetBool("global")
		
		// Determine if this should be a global or local config change
		global := true
		if len(args) > 1 && !globalFlag {
			// Path provided and --global not set = local config
			repoPath = args[1]
			global = false

			// Check if path is a git repository
			if !isGitRepository(repoPath) {
				return fmt.Errorf("❌ '%s' is not a git repository", repoPath)
			}
		} else if len(args) > 1 && globalFlag {
			// Both path and --global provided = error
			return fmt.Errorf("❌ Cannot specify both a path and --global flag. Use either 'krakn use %s %s' OR 'krakn use %s --global'", accountName, args[1], accountName)
		}
		// If globalFlag is true or no path provided = global config

		// Load config
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Find account
		account := config.getAccount(accountName)
		if account == nil {
			if len(config.Accounts) == 0 {
				return fmt.Errorf("❌ No accounts configured. Use 'krakn add' to add accounts first")
			}
			
			var availableNames []string
			for _, acc := range config.Accounts {
				availableNames = append(availableNames, acc.Name)
			}
			return fmt.Errorf("❌ Account '%s' not found. Available accounts: %s", accountName, strings.Join(availableNames, ", "))
		}

		// Update git config
		if err := setGitConfig("user.name", account.Username, repoPath, global); err != nil {
			return fmt.Errorf("failed to set git user.name: %w", err)
		}

		if err := setGitConfig("user.email", account.Email, repoPath, global); err != nil {
			return fmt.Errorf("failed to set git user.email: %w", err)
		}

		// Update current account in config
		if global {
			config.CurrentAccount = accountName
			if err := config.saveConfig(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
		}

		// Display success message
		scope := "globally"
		if !global {
			scope = fmt.Sprintf("for repository at %s", repoPath)
		} else if globalFlag {
			scope = "globally (via --global flag)"
		}

		fmt.Printf("🎉 Successfully using: %s\n", accountName)
		fmt.Printf("✅ Switched to account '%s' %s\n", accountName, scope)
		fmt.Printf("👤 Name: %s\n", account.Username)
		fmt.Printf("📧 Email: %s\n", account.Email)
		fmt.Printf("🔗 SSH Host: github.com-%s\n", accountName)

		if !global {
			fmt.Printf("\n💡 To clone repositories with this account, use:\n")
			fmt.Printf("   git clone git@github.com-%s:username/repo.git\n", accountName)
		} else {
			fmt.Printf("\n💡 Global git configuration updated!\n")
			fmt.Printf("   All new repositories will use this account by default\n")
		}

		return nil
	},
}

func isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if info, err := os.Stat(gitDir); err == nil {
		return info.IsDir()
	}
	return false
}

func setGitConfig(key, value, repoPath string, global bool) error {
	var cmd *exec.Cmd

	if global {
		cmd = exec.Command("git", "config", "--global", key, value)
	} else {
		cmd = exec.Command("git", "-C", repoPath, "config", key, value)
	}

	return cmd.Run()
}

func init() {
	RootCmd.AddCommand(useCmd)
	
	// Add the --global flag
	useCmd.Flags().BoolP("global", "g", false, "Set global git configuration (default behavior when no path is provided)")
}
