package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Account struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	SSHKey    string `json:"ssh_key"`
	Username  string `json:"username"`
	IsDefault bool   `json:"is_default"`
}

type Config struct {
	Accounts        []Account `json:"accounts"`
	CurrentAccount  string    `json:"current_account"`
	MigrationDone   bool      `json:"migration_done"`
}

func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".krakncat", "config.json")
}

func ensureConfigDir() error {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".krakncat")
	return os.MkdirAll(configDir, 0755)
}

func loadConfig() (*Config, error) {
	configPath := getConfigPath()
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Accounts:       []Account{},
			CurrentAccount: "",
			MigrationDone:  false,
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

func (c *Config) saveConfig() error {
	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := getConfigPath()
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (c *Config) addAccount(account Account) error {
	// Check if account already exists
	for i, existing := range c.Accounts {
		if existing.Name == account.Name {
			c.Accounts[i] = account
			return c.saveConfig()
		}
	}

	// Add new account
	c.Accounts = append(c.Accounts, account)
	
	// Set as default if it's the first account
	if len(c.Accounts) == 1 {
		account.IsDefault = true
		c.CurrentAccount = account.Name
		c.Accounts[0] = account
	}

	return c.saveConfig()
}

func (c *Config) getAccount(name string) *Account {
	for _, account := range c.Accounts {
		if account.Name == name {
			return &account
		}
	}
	return nil
}

func (c *Config) setCurrentAccount(name string) error {
	account := c.getAccount(name)
	if account == nil {
		return fmt.Errorf("account '%s' not found", name)
	}

	c.CurrentAccount = name
	return c.saveConfig()
}
