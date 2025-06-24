package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Provider represents a Git hosting provider
type Provider struct {
	Name         string `json:"name"`          // "github", "gitlab", "gitea", "custom"
	DisplayName  string `json:"display_name"` // "GitHub", "GitLab", "Gitea"
	Hostname     string `json:"hostname"`     // "github.com", "gitlab.com", "git.company.com"
	SSHUser      string `json:"ssh_user"`     // Usually "git"
	SSHPort      string `json:"ssh_port,omitempty"` // SSH port, empty for default (22)
	WebURL       string `json:"web_url"`      // For SSH key management URL
	KeySuffix    string `json:"key_suffix"`   // "gh", "gl", "gitea"
}

// Account represents a user account on a specific provider
type AccountV2 struct {
	Name         string   `json:"name"`          // "personal", "work"
	Email        string   `json:"email"`
	SSHKey       string   `json:"ssh_key"`
	Username     string   `json:"username"`
	Provider     Provider `json:"provider"`
	IsDefault    bool     `json:"is_default"`
}

type ConfigV2 struct {
	Accounts        []AccountV2 `json:"accounts"`
	CurrentAccount  string      `json:"current_account"`
	MigrationDone   bool        `json:"migration_done"`
	ConfigVersion   int         `json:"config_version"` // For future migrations
}

// Predefined providers
var DefaultProviders = map[string]Provider{
	"github": {
		Name:        "github",
		DisplayName: "GitHub",
		Hostname:    "github.com",
		SSHUser:     "git",
		WebURL:      "https://github.com/settings/ssh/new",
		KeySuffix:   "gh",
	},
	"gitlab": {
		Name:        "gitlab",
		DisplayName: "GitLab",
		Hostname:    "gitlab.com",
		SSHUser:     "git",
		WebURL:      "https://gitlab.com/-/profile/keys",
		KeySuffix:   "gl",
	},
	"gitea": {
		Name:        "gitea",
		DisplayName: "Gitea",
		Hostname:    "gitea.com", // Can be overridden for self-hosted
		SSHUser:     "git",
		WebURL:      "https://gitea.com/user/settings/keys",
		KeySuffix:   "gitea",
	},
}

// Helper functions for the new multi-provider system

func (a *AccountV2) GetSSHHost() string {
	return fmt.Sprintf("%s-%s", a.Provider.Hostname, a.Name)
}

func (a *AccountV2) GetSSHCloneURL(repo string) string {
	return fmt.Sprintf("git@%s:%s", a.GetSSHHost(), repo)
}

func (a *AccountV2) GetKeyPath() string {
	if a.SSHKey != "" {
		return a.SSHKey
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_ed25519_%s_%s", a.Provider.KeySuffix, a.Name))
}

func (a *AccountV2) GenerateSSHConfig() string {
	config := fmt.Sprintf(`
Host %s
  HostName %s
  User %s
  IdentityFile %s`, a.GetSSHHost(), a.Provider.Hostname, a.Provider.SSHUser, a.GetKeyPath())

	// Add port if not default
	if a.Provider.SSHPort != "" && a.Provider.SSHPort != "22" {
		config += fmt.Sprintf("\n  Port %s", a.Provider.SSHPort)
	}
	
	config += "\n"
	return config
}

// Migration function to convert old config to new format
func migrateConfigToV2(oldConfig *Config) *ConfigV2 {
	newConfig := &ConfigV2{
		Accounts:       []AccountV2{},
		CurrentAccount: oldConfig.CurrentAccount,
		MigrationDone:  oldConfig.MigrationDone,
		ConfigVersion:  2,
	}

	// Convert old accounts to new format (assume GitHub)
	for _, oldAccount := range oldConfig.Accounts {
		newAccount := AccountV2{
			Name:      oldAccount.Name,
			Email:     oldAccount.Email,
			SSHKey:    oldAccount.SSHKey,
			Username:  oldAccount.Username,
			Provider:  DefaultProviders["github"], // Default to GitHub
			IsDefault: oldAccount.IsDefault,
		}
		newConfig.Accounts = append(newConfig.Accounts, newAccount)
	}

	return newConfig
}

// Enhanced provider detection from SSH config
func detectProviderFromSSHHost(host string) *Provider {
	if strings.Contains(host, "github.com") {
		provider := DefaultProviders["github"]
		return &provider
	}
	if strings.Contains(host, "gitlab.com") {
		provider := DefaultProviders["gitlab"]
		return &provider
	}
	if strings.Contains(host, "gitea") {
		provider := DefaultProviders["gitea"]
		return &provider
	}
	
	// For custom/self-hosted instances
	parts := strings.Split(host, "-")
	if len(parts) >= 2 {
		hostname := strings.Join(parts[:len(parts)-1], "-")
		return &Provider{
			Name:        "custom",
			DisplayName: "Custom Git Host",
			Hostname:    hostname,
			SSHUser:     "git",
			WebURL:      fmt.Sprintf("https://%s", hostname),
			KeySuffix:   "custom",
		}
	}
	
	return nil
}

// Interactive provider selection
func selectProvider() (*Provider, error) {
	fmt.Println("\n🌐 Select Git hosting provider:")
	fmt.Println("   1. GitHub (github.com)")
	fmt.Println("   2. GitLab (gitlab.com)")  
	fmt.Println("   3. Gitea (gitea.com)")
	fmt.Println("   4. Custom/Self-hosted (e.g., git.company.com, code.myorg.io)")
	
	var choice int
	fmt.Print("Enter choice (1-4): ")
	if _, err := fmt.Scanf("%d", &choice); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	
	switch choice {
	case 1:
		provider := DefaultProviders["github"]
		return &provider, nil
	case 2:
		provider := DefaultProviders["gitlab"]
		return &provider, nil
	case 3:
		provider := DefaultProviders["gitea"]
		return &provider, nil
	case 4:
		return createCustomProvider()
	default:
		return nil, fmt.Errorf("invalid choice")
	}
}

func createCustomProvider() (*Provider, error) {
	fmt.Println("\n🔧 Custom Git Provider Setup")
	fmt.Println("   Configure your self-hosted Git server or custom Git hosting")
	
	// Get hostname
	fmt.Print("\n🌐 Enter hostname (e.g., git.company.com, code.myorg.io): ")
	var hostname string
	if _, err := fmt.Scanf("%s", &hostname); err != nil {
		return nil, fmt.Errorf("invalid hostname: %w", err)
	}
	
	// Validate hostname format
	if !isValidHostname(hostname) {
		return nil, fmt.Errorf("invalid hostname format: %s", hostname)
	}
	
	// Get display name
	fmt.Printf("📝 Enter display name [%s]: ", hostname)
	var displayName string
	if _, err := fmt.Scanf("%s", &displayName); err != nil || displayName == "" {
		displayName = hostname // fallback to hostname
	}
	
	// Get SSH user (default: git)
	fmt.Print("👤 SSH user [git]: ")
	var sshUser string
	if _, err := fmt.Scanf("%s", &sshUser); err != nil || sshUser == "" {
		sshUser = "git"
	}
	
	// Get SSH port if non-standard
	fmt.Print("🔌 SSH port [22]: ")
	var port string
	if _, err := fmt.Scanf("%s", &port); err != nil || port == "" {
		port = "22"
	}
	
	// Ask about SSH key management URL
	fmt.Printf("🔗 SSH key management URL [https://%s]: ", hostname)
	var webURL string
	if _, err := fmt.Scanf("%s", &webURL); err != nil || webURL == "" {
		webURL = fmt.Sprintf("https://%s", hostname)
	}
	
	// Generate key suffix from hostname
	keySuffix := generateKeySuffix(hostname)
	fmt.Printf("🔑 SSH key suffix will be: %s\n", keySuffix)
	
	// Create provider
	provider := &Provider{
		Name:        "custom",
		DisplayName: displayName,
		Hostname:    hostname,
		SSHUser:     sshUser,
		WebURL:      webURL,
		KeySuffix:   keySuffix,
	}
	
	// Add SSH port if non-standard
	if port != "22" {
		provider.SSHPort = port
	}
	
	// Confirm configuration
	fmt.Println("\n✅ Custom provider configuration:")
	fmt.Printf("   Name: %s\n", provider.DisplayName)
	fmt.Printf("   Hostname: %s\n", provider.Hostname)
	fmt.Printf("   SSH User: %s\n", provider.SSHUser)
	if provider.SSHPort != "" {
		fmt.Printf("   SSH Port: %s\n", provider.SSHPort)
	}
	fmt.Printf("   Web URL: %s\n", provider.WebURL)
	fmt.Printf("   Key Suffix: %s\n", provider.KeySuffix)
	
	fmt.Print("\n💾 Save this configuration? [Y/n]: ")
	var confirm string
	if _, err := fmt.Scanf("%s", &confirm); err != nil {
		confirm = "y"
	}
	
	if strings.ToLower(confirm) == "n" {
		return nil, fmt.Errorf("configuration cancelled")
	}
	
	return provider, nil
}

// Helper functions for custom provider validation and configuration

// isValidHostname validates if a hostname is properly formatted
func isValidHostname(hostname string) bool {
	if hostname == "" {
		return false
	}
	
	// Basic hostname validation
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-\.]*[a-zA-Z0-9])?$`)
	if !hostnameRegex.MatchString(hostname) {
		return false
	}
	
	// Check if it's a valid URL-like hostname
	if strings.Contains(hostname, "://") {
		if _, err := url.Parse(hostname); err != nil {
			return false
		}
	}
	
	return true
}

// generateKeySuffix creates a short suffix for SSH key naming from hostname
func generateKeySuffix(hostname string) string {
	// Remove common prefixes
	hostname = strings.TrimPrefix(hostname, "git.")
	hostname = strings.TrimPrefix(hostname, "code.")
	hostname = strings.TrimPrefix(hostname, "source.")
	
	// Split by dots and take meaningful parts
	parts := strings.Split(hostname, ".")
	if len(parts) >= 2 {
		// Use first part of domain (e.g., "company" from "company.com")
		suffix := parts[0]
		// Truncate if too long
		if len(suffix) > 8 {
			suffix = suffix[:8]
		}
		return suffix
	}
	
	// Fallback: use first 8 characters
	if len(hostname) > 8 {
		return hostname[:8]
	}
	
	return hostname
}
