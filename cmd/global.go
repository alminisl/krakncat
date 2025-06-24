package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var globalCmd = &cobra.Command{
	Use:   "global [account-name]",
	Short: "Set global git configuration to use a specific account",
	Long: `Set the global git configuration to use a specific account.
This updates ~/.gitconfig with the default user.name and user.email.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountName := args[0]

		// Load config to get account details
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

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

		// Set global git config
		if err := setGlobalGitConfig("user.name", account.Username); err != nil {
			return fmt.Errorf("failed to set global git user.name: %w", err)
		}

		if err := setGlobalGitConfig("user.email", account.Email); err != nil {
			return fmt.Errorf("failed to set global git user.email: %w", err)
		}

		// Update current account in config
		config.CurrentAccount = accountName
		if err := config.saveConfig(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✅ Global git configuration set to account '%s'\n", accountName)
		fmt.Printf("👤 Name: %s\n", account.Username)
		fmt.Printf("📧 Email: %s\n", account.Email)
		fmt.Printf("🔗 SSH Host: github.com-%s\n", accountName)
		fmt.Println("\n💡 This will be used as the default for all repositories unless overridden by conditional includes!")

		return nil
	},
}

var showIncludesCmd = &cobra.Command{
	Use:   "show-includes",
	Short: "Show current conditional includes in global git config",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, _ := os.UserHomeDir()
		globalConfigPath := filepath.Join(homeDir, ".gitconfig")

		// Read global .gitconfig
		content, err := os.ReadFile(globalConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read global .gitconfig: %w", err)
		}

		configStr := string(content)
		fmt.Println("🔧 Global Git Configuration:")
		fmt.Printf("📁 File: %s\n\n", globalConfigPath)

		// Parse and display conditional includes
		lines := strings.Split(configStr, "\n")
		var inIncludeSection bool
		hasIncludes := false

		fmt.Println("📋 Conditional Includes:")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			
			if strings.HasPrefix(line, "[includeIf") {
				inIncludeSection = true
				hasIncludes = true
				// Extract the gitdir pattern
				start := strings.Index(line, "\"gitdir:")
				end := strings.Index(line[start+8:], "\"")
				if start != -1 && end != -1 {
					gitdir := line[start+8 : start+8+end]
					fmt.Printf("  📁 %s\n", gitdir)
				}
			} else if inIncludeSection && strings.HasPrefix(line, "path") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					path := strings.TrimSpace(parts[1])
					fmt.Printf("    🔗 → %s\n", path)
				}
				inIncludeSection = false
			} else if strings.HasPrefix(line, "[") {
				inIncludeSection = false
			}
		}

		if !hasIncludes {
			fmt.Println("  ℹ️  No conditional includes configured yet")
			fmt.Println("  💡 Use 'krakn setup-dir' or 'krakn config-dir' to create them")
		}

		return nil
	},
}

func setGlobalGitConfig(key, value string) error {
	cmd := exec.Command("git", "config", "--global", key, value)
	return cmd.Run()
}

func init() {
	RootCmd.AddCommand(globalCmd)
	RootCmd.AddCommand(showIncludesCmd)
}
