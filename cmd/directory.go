package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var dirConfigCmd = &cobra.Command{
	Use:   "config [directory] [account-name]",
	Short: "Setup automatic git config for a directory using conditional includes",
	Long: `Setup automatic git configuration for a directory using Git's conditional includes.
If no arguments provided, interactively configures the current directory.
If directory and account provided, configures that directory for the account.

Examples:
  krakn config                     # Interactive setup for current directory
  krakn config ~/work personal     # Setup ~/work for 'personal' account
  krakn config . work              # Setup current directory for 'work' account`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Interactive mode (no arguments)
		if len(args) == 0 {
			return interactiveDirectoryConfig()
		}

		// Direct mode (directory and account provided)
		if len(args) != 2 {
			return fmt.Errorf("provide either no arguments (interactive) or both directory and account-name")
		}

		dirPath := args[0]
		accountName := args[1]

		// Resolve absolute path
		absPath, err := filepath.Abs(dirPath)
		if err != nil {
			return fmt.Errorf("failed to resolve directory path: %w", err)
		}

		// Ensure directory exists
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

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
			return fmt.Errorf("❌ Account '%s' not found. Available accounts: %s", 
				accountName, strings.Join(availableNames, ", "))
		}

		return setupDirectoryConfig(absPath, account)
	},
}

func interactiveDirectoryConfig() error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	// Load available accounts
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(config.Accounts) == 0 {
		return fmt.Errorf("❌ No accounts configured. Use 'krakn add' to add accounts first")
	}

	// Show current directory
	fmt.Printf("📁 Current directory: %s\n\n", currentDir)

	// Show available accounts
	fmt.Println("📋 Available accounts:")
	for i, account := range config.Accounts {
		fmt.Printf("  %d. %s (%s)\n", i+1, account.Name, account.Email)
	}

	// Ask user to select account
	fmt.Print("\n💬 Select account number: ")
	resp, _ := reader.ReadString('\n')
	resp = strings.TrimSpace(resp)

	// Parse selection
	var selectedAccount *Account
	for i, account := range config.Accounts {
		if resp == fmt.Sprintf("%d", i+1) {
			selectedAccount = &account
			break
		}
	}

	if selectedAccount == nil {
		return fmt.Errorf("❌ Invalid selection")
	}

	// Setup the directory
	return setupDirectoryConfig(currentDir, selectedAccount)
}

func addConditionalInclude(dirPath, configPath string) error {
	homeDir, _ := os.UserHomeDir()
	globalConfigPath := filepath.Join(homeDir, ".gitconfig")

	// Prepare the conditional include entry
	// Git requires trailing slash for gitdir
	gitDirPattern := dirPath
	if !strings.HasSuffix(gitDirPattern, "/") {
		gitDirPattern += "/"
	}

	includeSection := fmt.Sprintf("\n[includeIf \"gitdir:%s\"]\n\tpath = %s\n", gitDirPattern, configPath)

	// Check if this include already exists
	if existingConfig, err := os.ReadFile(globalConfigPath); err == nil {
		if strings.Contains(string(existingConfig), fmt.Sprintf("gitdir:%s", gitDirPattern)) {
			fmt.Println("ℹ️  Conditional include already exists in global .gitconfig")
			return nil
		}
	}

	// Append to global .gitconfig
	f, err := os.OpenFile(globalConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open global .gitconfig: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(includeSection); err != nil {
		return fmt.Errorf("failed to write conditional include: %w", err)
	}

	fmt.Println("✅ Added conditional include to global .gitconfig")
	return nil
}

func setupDirectoryConfig(dirPath string, account *Account) error {
	// Create directory-specific .gitconfig
	gitConfigPath := filepath.Join(dirPath, ".gitconfig")
	gitConfigContent := fmt.Sprintf(`[user]
	name = %s
	email = %s
`, account.Username, account.Email)

	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitconfig: %w", err)
	}

	// Add conditional include to global .gitconfig
	if err := addConditionalInclude(dirPath, gitConfigPath); err != nil {
		return fmt.Errorf("failed to add conditional include: %w", err)
	}

	fmt.Printf("✅ Directory '%s' configured for account '%s'\n", dirPath, account.Name)
	fmt.Printf("👤 Name: %s\n", account.Username)
	fmt.Printf("📧 Email: %s\n", account.Email)
	fmt.Printf("📁 Config file: %s\n", gitConfigPath)
	fmt.Printf("🔗 SSH Host: github.com-%s\n", account.Name)
	fmt.Println("\n💡 Git will automatically use these settings in this directory!")

	return nil
}

func init() {
	RootCmd.AddCommand(dirConfigCmd)
}
