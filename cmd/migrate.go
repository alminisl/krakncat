package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// DiscoveredAccount represents a potential account found during migration
type DiscoveredAccount struct {
	Name       string
	Email      string
	Username   string
	Source     string // "global", "ssh-config", etc.
	Suggested  bool   // Whether this is a suggested match
}

// checkAndOfferMigration checks if this is first run and offers to migrate existing git config
func checkAndOfferMigration() error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	// Skip if migration already done or accounts already exist
	if config.MigrationDone || len(config.Accounts) > 0 {
		return nil
	}

	// Discover all potential accounts
	discovered := discoverExistingAccounts()

	if len(discovered) == 0 {
		// No existing configuration found, mark migration as done
		config.MigrationDone = true
		return config.saveConfig()
	}

	// Offer migration
	fmt.Println("👋 Welcome to krakncat!")
	fmt.Println("\n🔍 I found existing git/SSH configuration:")

	for i, acc := range discovered {
		fmt.Printf("\n   %d. %s", i+1, acc.Source)
		if acc.Name != "" {
			fmt.Printf(" - Name: %s", acc.Name)
		}
		if acc.Email != "" {
			fmt.Printf(" - Email: %s", acc.Email)
		}
		if acc.Username != "" {
			fmt.Printf(" - Username: %s", acc.Username)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n💫 Would you like to migrate any of these accounts to krakncat? [Y/n]: ")
	resp, _ := reader.ReadString('\n')
	resp = strings.ToLower(strings.TrimSpace(resp))

	if resp == "n" || resp == "no" {
		config.MigrationDone = true
		return config.saveConfig()
	}

	// Let user select which accounts to migrate
	selected := selectAccountsToMigrate(discovered)
	if len(selected) == 0 {
		config.MigrationDone = true
		return config.saveConfig()
	}

	// Migrate selected accounts
	for _, acc := range selected {
		migratedAccount, err := migrateAccount(acc)
		if err != nil {
			fmt.Printf("❌ Failed to migrate account: %v\n", err)
			continue
		}
		config.Accounts = append(config.Accounts, migratedAccount)
	}

	// Set first account as current
	if len(config.Accounts) > 0 {
		config.CurrentAccount = config.Accounts[0].Name
	}

	config.MigrationDone = true

	if err := config.saveConfig(); err != nil {
		return fmt.Errorf("failed to save migrated config: %w", err)
	}

	fmt.Printf("\n✅ Successfully imported %d account(s)!\n", len(config.Accounts))
	
	fmt.Printf("\n🎯 Next steps:\n")
	fmt.Printf("   • Use 'krakn list' to see your accounts\n")
	fmt.Printf("   • Use 'krakn config ~/work work' to set up directory-based switching\n")
	fmt.Printf("   • Use 'krakn add' to add more accounts\n")

	return nil
}

// discoverExistingAccounts looks for existing git config and SSH configuration
func discoverExistingAccounts() []DiscoveredAccount {
	var discovered []DiscoveredAccount

	// Check global git config
	globalUser := getGitConfigValue("user.name", true)
	globalEmail := getGitConfigValue("user.email", true)

	if globalUser != "" || globalEmail != "" {
		discovered = append(discovered, DiscoveredAccount{
			Name:     globalUser,
			Email:    globalEmail,
			Source:   "Global Git Config",
			Suggested: true,
		})
	}

	// Check SSH config for existing GitHub hosts
	sshAccounts := discoverSSHAccounts()
	discovered = append(discovered, sshAccounts...)

	return discovered
}

// discoverSSHAccounts parses ~/.ssh/config for existing GitHub account configurations
func discoverSSHAccounts() []DiscoveredAccount {
	var accounts []DiscoveredAccount

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return accounts
	}

	sshConfigPath := filepath.Join(homeDir, ".ssh", "config")
	content, err := os.ReadFile(sshConfigPath)
	if err != nil {
		return accounts
	}

	lines := strings.Split(string(content), "\n")
	var currentHost string
	var currentUser string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Match Host github.com-* patterns
		if strings.HasPrefix(line, "Host github.com-") {
			currentHost = strings.TrimPrefix(line, "Host ")
			currentUser = ""
		} else if strings.HasPrefix(line, "User ") && currentHost != "" {
			currentUser = strings.TrimPrefix(line, "User ")
			
			// Extract account name from host
			accountName := strings.TrimPrefix(currentHost, "github.com-")
			if accountName != "" && accountName != "github.com" {
				accounts = append(accounts, DiscoveredAccount{
					Username: currentUser,
					Source:   fmt.Sprintf("SSH Config (%s)", currentHost),
					Suggested: false,
				})
			}
		}
	}

	return accounts
}

// selectAccountsToMigrate lets the user choose which accounts to migrate
func selectAccountsToMigrate(discovered []DiscoveredAccount) []DiscoveredAccount {
	reader := bufio.NewReader(os.Stdin)
	var selected []DiscoveredAccount

	fmt.Println("\n📋 Select accounts to migrate:")
	fmt.Println("   0. Skip migration")

	for i, acc := range discovered {
		fmt.Printf("   %d. %s", i+1, acc.Source)
		if acc.Suggested {
			fmt.Print(" (recommended)")
		}
		fmt.Println()
	}

	fmt.Printf("   %d. Migrate all\n", len(discovered)+1)

	for {
		fmt.Print("\nEnter your choice(s) separated by commas (e.g., 1,3): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "0" {
			return selected
		}

		if input == strconv.Itoa(len(discovered)+1) {
			return discovered
		}

		// Parse selection
		choices := strings.Split(input, ",")
		selected = []DiscoveredAccount{}
		valid := true

		for _, choice := range choices {
			choice = strings.TrimSpace(choice)
			index, err := strconv.Atoi(choice)
			if err != nil || index < 1 || index > len(discovered) {
				fmt.Printf("❌ Invalid choice: %s\n", choice)
				valid = false
				break
			}
			selected = append(selected, discovered[index-1])
		}

		if valid {
			break
		}
	}

	return selected
}

// migrateAccount migrates a single discovered account
func migrateAccount(discovered DiscoveredAccount) (Account, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n🔧 Migrating: %s\n", discovered.Source)

	// Get account name
	var accountName string
	if discovered.Username != "" {
		fmt.Printf("📝 Account name [%s]: ", discovered.Username)
	} else {
		fmt.Print("📝 Account name (e.g., 'personal', 'work'): ")
	}
	
	input, _ := reader.ReadString('\n')
	accountName = strings.TrimSpace(input)
	
	if accountName == "" {
		if discovered.Username != "" {
			accountName = discovered.Username
		} else {
			accountName = "default"
		}
	}

	// Get email if not provided
	email := discovered.Email
	if email == "" {
		fmt.Print("📧 Email address: ")
		input, _ := reader.ReadString('\n')
		email = strings.TrimSpace(input)
	}

	// Get GitHub username if not provided
	username := discovered.Username
	if username == "" {
		if discovered.Name != "" {
			fmt.Printf("👤 GitHub username [%s]: ", discovered.Name)
		} else {
			fmt.Print("👤 GitHub username: ")
		}
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		if input != "" {
			username = input
		} else if discovered.Name != "" {
			username = discovered.Name
		}
	}

	// Select SSH key
	sshKey := selectSSHKey(accountName)

	account := Account{
		Name:     accountName,
		Email:    email,
		SSHKey:   sshKey,
		Username: username,
	}

	fmt.Printf("✅ Configured account '%s'\n", accountName)
	fmt.Printf("   � Email: %s\n", email)
	fmt.Printf("   👤 Username: %s\n", username)
	fmt.Printf("   �🔗 SSH Host: github.com-%s\n", accountName)
	if sshKey != "" {
		fmt.Printf("   🔑 SSH Key: %s\n", sshKey)
	}

	return account, nil
}

// selectSSHKey helps user select or specify an SSH key for the account
func selectSSHKey(accountName string) string {
	reader := bufio.NewReader(os.Stdin)
	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")

	// Find existing SSH keys
	var existingKeys []string
	if entries, err := os.ReadDir(sshDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasSuffix(entry.Name(), ".pub") {
				// Check if it's a private key (has corresponding .pub file)
				pubKeyPath := filepath.Join(sshDir, entry.Name()+".pub")
				if _, err := os.Stat(pubKeyPath); err == nil {
					existingKeys = append(existingKeys, entry.Name())
				}
			}
		}
	}

	if len(existingKeys) == 0 {
		fmt.Println("🔑 No existing SSH keys found.")
		fmt.Print("   SSH key path (leave empty to generate later): ")
		input, _ := reader.ReadString('\n')
		return strings.TrimSpace(input)
	}

	fmt.Println("\n🔑 SSH Key Options:")
	fmt.Println("   0. Generate new key later")
	
	for i, key := range existingKeys {
		fmt.Printf("   %d. %s", i+1, key)
		// Highlight suggested key
		if strings.Contains(key, accountName) || strings.Contains(key, "ed25519") {
			fmt.Print(" (suggested)")
		}
		fmt.Println()
	}

	fmt.Print("   Enter custom path\n")

	for {
		fmt.Print("\nSelect SSH key [0]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" || input == "0" {
			return ""
		}

		// Check if it's a number selection
		if index, err := strconv.Atoi(input); err == nil {
			if index > 0 && index <= len(existingKeys) {
				return filepath.Join(sshDir, existingKeys[index-1])
			}
			fmt.Printf("❌ Invalid selection: %d\n", index)
			continue
		}

		// Treat as custom path
		if strings.HasPrefix(input, "~/") {
			input = filepath.Join(homeDir, input[2:])
		}

		// Verify the key exists
		if _, err := os.Stat(input); err == nil {
			return input
		}

		fmt.Printf("❌ SSH key not found: %s\n", input)
	}
}

func getGitConfigValue(key string, global bool) string {
	var cmd *exec.Cmd
	if global {
		cmd = exec.Command("git", "config", "--global", "--get", key)
	} else {
		cmd = exec.Command("git", "config", "--get", key)
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate existing git configuration to krakncat",
	Long: `Migrate your existing global git configuration to krakncat.
This command helps you import your current git user.name and user.email
as your first krakncat account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Force migration even if already done
		config, err := loadConfig()
		if err != nil {
			return err
		}

		config.MigrationDone = false
		if err := config.saveConfig(); err != nil {
			return err
		}

		return checkAndOfferMigration()
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}
