package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// generateSSHKey generates an SSH key and optionally updates SSH config
func generateSSHKey(name, email, keyPath string) error {
	// Ensure the SSH directory exists
	if err := ensureSSHDirectory(); err != nil {
		return err
	}

	// Ensure the key directory exists
	if err := ensureSSHKeyDirectory(keyPath); err != nil {
		return err
	}

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		return fmt.Errorf("❌ SSH key already exists at %s", keyPath)
	}

	// Generate SSH key
	cmdArgs := []string{
		"-t", "ed25519",
		"-C", email,
		"-f", keyPath,
		"-q",
		"-N", "",
	}

	cmdGen := exec.Command("ssh-keygen", cmdArgs...)
	cmdGen.Stdin = os.Stdin
	cmdGen.Stdout = os.Stdout
	cmdGen.Stderr = os.Stderr

	if err := cmdGen.Run(); err != nil {
		return fmt.Errorf("failed to generate ssh key: %w", err)
	}

	// Read public key
	pubKey, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		return fmt.Errorf("could not read public key: %w", err)
	}

	// Create SSH config snippet
	sshConfigSnippet := fmt.Sprintf(`

Host github.com-%s
  HostName github.com
  User git
  IdentityFile %s
`, name, keyPath)

	// Ask user if they want to update SSH config
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n💬 Do you want to append this config to ~/.ssh/config? [Y/n]: ")
	resp, _ := reader.ReadString('\n')
	resp = strings.ToLower(strings.TrimSpace(resp))

	if resp == "y" || resp == "" {
		// Ensure SSH directory exists before writing config
		if err := ensureSSHDirectory(); err != nil {
			return err
		}
		
		homeDir, _ := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".ssh", "config")

		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to write to SSH config: %w", err)
		}
		defer f.Close()
		if _, err := f.WriteString(sshConfigSnippet); err != nil {
			return fmt.Errorf("could not write config: %w", err)
		}
		fmt.Println("✅ SSH config updated.")
	} else {
		fmt.Println("⚠️ Skipped modifying ~/.ssh/config.")
	}

	fmt.Println("\n✅ SSH key created at:", keyPath)
	fmt.Println("\n🔑 Public key:\n" + string(pubKey))
	fmt.Println("\n📋 Add this public key to GitHub: https://github.com/settings/ssh/new")
	fmt.Printf("🌐 Host alias for SSH: github.com-%s\n", name)

	return nil
}

// ensureSSHDirectory creates the .ssh directory if it doesn't exist with proper permissions
func ensureSSHDirectory() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	return nil
}

// ensureSSHKeyDirectory creates the directory for an SSH key path if it doesn't exist
func ensureSSHKeyDirectory(keyPath string) error {
	keyDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create directory for SSH key %s: %w", keyDir, err)
	}
	return nil
}

var generateKeyCmd = &cobra.Command{
	Use:   "generate-key",
	Short: "Generate and configure a new SSH key for a GitHub account",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		email, _ := cmd.Flags().GetString("email")

		if name == "" || email == "" {
			return fmt.Errorf("please provide both --name and --email")
		}

		homeDir, _ := os.UserHomeDir()
		sshDir := filepath.Join(homeDir, ".ssh")
		keyPath := filepath.Join(sshDir, fmt.Sprintf("id_ed25519_gh_%s", name))

		if err := generateSSHKey(name, email, keyPath); err != nil {
			return err
		}

		// Ask if user wants to save account configuration
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\n💾 Do you want to save this as an account configuration? [Y/n]: ")
		resp, _ := reader.ReadString('\n')
		resp = strings.ToLower(strings.TrimSpace(resp))

		if resp == "y" || resp == "" {
			fmt.Print("👤 GitHub username: ")
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)

			if username != "" {
				config, err := loadConfig()
				if err != nil {
					fmt.Printf("⚠️  Could not load config: %v\n", err)
					return nil
				}

				account := Account{
					Name:     name,
					Email:    email,
					SSHKey:   keyPath,
					Username: username,
				}

				if err := config.addAccount(account); err != nil {
					fmt.Printf("⚠️  Could not save account: %v\n", err)
					return nil
				}

				fmt.Printf("✅ Account '%s' saved to configuration!\n", name)
			}
		}

		return nil
	},
}

func init() {
	generateKeyCmd.Flags().String("name", "", "Unique account name (e.g. 'work')")
	generateKeyCmd.Flags().String("email", "", "Email address for SSH key")
	RootCmd.AddCommand(generateKeyCmd)
}
