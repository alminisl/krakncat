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

var generateKeyCmd = &cobra.Command{
	Use:   "generate-key",
	Short: "Generate and configure a new SSH key for a GitHub account",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		email, _ := cmd.Flags().GetString("email")

		if name == "" || email == "" {
			return fmt.Errorf("please provide both --name and --email")
		}

		sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")
		keyPath := filepath.Join(sshDir, fmt.Sprintf("id_ed25519_gh_%s", name))

		if _, err := os.Stat(keyPath); err == nil {
			return fmt.Errorf("❌ SSH key already exists at %s", keyPath)
		}

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

		pubKey, err := os.ReadFile(keyPath + ".pub")
		if err != nil {
			return fmt.Errorf("could not read public key: %w", err)
		}

		sshConfigSnippet := fmt.Sprintf(`

Host github.com-%s
  HostName github.com
  User git
  IdentityFile %s
`, name, keyPath)

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("\n💬 Do you want to append this config to ~/.ssh/config? [Y/n]: ")
		resp, _ := reader.ReadString('\n')
		resp = strings.ToLower(strings.TrimSpace(resp))

		if resp == "y" || resp == "" {
			configPath := filepath.Join(sshDir, "config")
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
	},
}

func init() {
	generateKeyCmd.Flags().String("name", "", "Unique account name (e.g. 'work')")
	generateKeyCmd.Flags().String("email", "", "Email address for SSH key")
	RootCmd.AddCommand(generateKeyCmd)
}
