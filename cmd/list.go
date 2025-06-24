package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all configured GitHub accounts",
	Long: `List all configured GitHub accounts and current git configuration.

Use --global flag to show only global git configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		globalOnly, _ := cmd.Flags().GetBool("global")

		if globalOnly {
			return showGlobalConfig()
		}

		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(config.Accounts) == 0 {
			fmt.Println("🚫 No accounts configured yet.")
			fmt.Println("💡 Use 'krakn add' to add your first account.")
			return nil
		}

		fmt.Println("📋 Configured GitHub accounts:")
		fmt.Println()

		for _, account := range config.Accounts {
			status := ""
			if account.Name == config.CurrentAccount {
				status = " ✅ (current)"
			}

			fmt.Printf("👤 %s%s\n", account.Name, status)
			fmt.Printf("   📧 Email: %s\n", account.Email)
			fmt.Printf("   🔑 SSH Key: %s\n", account.SSHKey)
			fmt.Printf("   🌐 GitHub: @%s\n", account.Username)
			fmt.Printf("   🔗 SSH Host: github.com-%s\n", account.Name)
			fmt.Println()
		}

		// Show current git config
		fmt.Println("🔧 Current Git Configuration:")
		if gitUser := getGitConfig("user.name", false); gitUser != "" {
			fmt.Printf("   👤 Name: %s\n", gitUser)
		} else {
			fmt.Println("   ℹ️  No local git configuration")
		}
		if gitEmail := getGitConfig("user.email", false); gitEmail != "" {
			fmt.Printf("   📧 Email: %s\n", gitEmail)
		}

		// Show global git config
		fmt.Println("\n🌍 Global Git Configuration:")
		if globalUser := getGitConfig("user.name", true); globalUser != "" {
			fmt.Printf("   👤 Name: %s\n", globalUser)
		} else {
			fmt.Println("   ℹ️  No global git user configured")
		}
		if globalEmail := getGitConfig("user.email", true); globalEmail != "" {
			fmt.Printf("   📧 Email: %s\n", globalEmail)
		}

		return nil
	},
}

func getGitConfig(key string, global bool) string {
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

func showGlobalConfig() error {
	fmt.Println("🌍 Global Git Configuration:")

	if globalUser := getGitConfig("user.name", true); globalUser != "" {
		fmt.Printf("   👤 Name: %s\n", globalUser)
	} else {
		fmt.Println("   ℹ️  No global git user configured")
	}

	if globalEmail := getGitConfig("user.email", true); globalEmail != "" {
		fmt.Printf("   📧 Email: %s\n", globalEmail)
	} else {
		fmt.Println("   ℹ️  No global git email configured")
	}

	return nil
}

func init() {
	listCmd.Flags().BoolP("global", "g", false, "Show only global git configuration")
	RootCmd.AddCommand(listCmd)
}
