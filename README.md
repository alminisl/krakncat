<div align="center">
  <img src="assets/krakncat.png" alt="krakncat Logo" width="200" height="200" />
</div>

# krakncat

🐙 A simple CLI tool to manage multiple GitHub accounts on your machine.

`krakncat` (or `krakn`) helps you generate SSH keys, configure git and SSH for each account, and easily switch between accounts when working with repositories.

## Features

- 🔑 **Generate SSH keys per account** with automatic SSH config setup
- 👥 **Add multiple GitHub accounts** with personalized configuration
- 📁 **Manage git user/email settings** per project directory
- 🔄 **Clone repositories** using the right SSH key and account
- 🎯 **Simple commands** with an intuitive CLI interface

## Installation

### Prerequisites

Ensure you have Go 1.21+ installed on your system.

**Arch Linux:**

```bash
sudo pacman -S go
```

**Other distributions:**

- Follow the [official Go installation guide](https://golang.org/doc/install)

### Build from source

```bash
git clone https://github.com/alminisl/krakncat.git
cd krakncat
make dev              # Download deps, tidy modules, and build
```

### Alternative build methods

**Using Go directly:**

```bash
go mod tidy
go build -o krakn .
```

**Using make:**

```bash
make build           # Just build
make install         # Build and install to /usr/local/bin
make test           # Run tests
make clean          # Clean build artifacts
```

### Installation to system

```bash
# After building
sudo cp krakn /usr/local/bin/
# or use make
make install
```

### Quick Install for Arch Linux 🏃‍♂️

```bash
git clone https://github.com/alminisl/krakncat.git
cd krakncat
./install-arch.sh    # Automated installation script
```

This script will:

- Install Go and other dependencies via pacman
- Build the application
- Optionally install it system-wide

### First Run - Smart Migration 🆕

When you run `krakncat` for the first time, it automatically detects ALL existing configurations and lets you choose what to import:

```bash
./krakn list
# 👋 Welcome to krakncat!
# 🔍 I found existing git/SSH configuration:
#
#    1. Global Git Config - Name: John Doe - Email: john@example.com (recommended)
#    2. SSH Config (github.com-work) - Username: john-work
#    3. SSH Config (github.com-personal) - Username: john-personal
#
# 💫 Would you like to migrate any of these accounts to krakncat? [Y/n]:
# 📋 Select accounts to migrate:
#    0. Skip migration
#    1. Global Git Config (recommended)
#    2. SSH Config (github.com-work)
#    3. SSH Config (github.com-personal)
#    4. Migrate all
#
# Enter your choice(s) separated by commas (e.g., 1,3): 1,2
```

**Smart Migration Features:**

- 🔍 **Auto-Detection**: Finds existing git config AND SSH configurations
- 🎯 **Multi-Account**: Detects multiple GitHub accounts from SSH config
- ✨ **Selective Import**: Choose exactly which accounts to migrate
- 🔑 **SSH Key Matching**: Automatically suggests appropriate SSH keys
- 📝 **Custom Naming**: Rename accounts during migration
- 🚀 **Zero Setup**: Creates complete account configurations instantly

### Requirements

- Go 1.24+
- Git
- SSH (ssh-keygen, ssh-agent)

## Usage

### Generate a new SSH key for an account

```bash
./krakn generate-key --name personal --email your.email@example.com
```

This command:

- Creates a new ED25519 SSH key pair
- Saves it as `~/.ssh/id_ed25519_gh_<name>`
- Optionally adds SSH config to `~/.ssh/config`
- Displays the public key for you to add to GitHub
- Optionally saves the account configuration for easy switching

### Add a new GitHub account

```bash
./krakn add
```

Interactive command that prompts for:

- Account name (e.g., 'work', 'personal')
- Email address
- GitHub username
- SSH key path (with option to generate if missing)

### List all accounts

```bash
./krakn list
# or
./krakn ls
```

Shows:

- All configured accounts with details
- Current active account (marked with ✅)
- Current git configuration (local and global)

### Switch accounts

```bash
# Switch globally
./krakn use personal

# Switch for a specific repository
./krakn use work /path/to/repo
```

This command:

- Updates git `user.name` and `user.email` configuration
- Works globally or for specific repositories
- Shows which SSH host to use for cloning

### Automatic Directory-Based Configuration (Git Conditional Includes)

🎯 **The most powerful feature!** Set up automatic account switching based on directory location.

```bash
# Set up automatic config for a directory
./krakn config ~/work/projects work
./krakn config ~/personal/projects personal

# Or configure current directory interactively
cd ~/work/my-project
./krakn config
```

This uses Git's conditional includes feature to automatically use the right account when you `cd` into different directories!

### Set Global Default

```bash
# Set global default account
./krakn global personal
```

### View Configuration

```bash
# Show current conditional includes
./krakn show-includes

# Show only global git config (quick check)
./krakn list --global
```

### Migration and Account Management

```bash
# Manually trigger migration of existing git/SSH config
./krakn migrate

# Remove an account when no longer needed
./krakn remove old-account
```

**Advanced Migration Features:**

- 🔍 **Smart Detection**: Automatically finds all GitHub configurations (git config + SSH hosts)
- 🎯 **Multi-Account Support**: Import multiple accounts in one session
- ✨ **Selective Migration**: Choose exactly which accounts to migrate
- 🔑 **SSH Key Selection**: Pick from existing keys or generate new ones
- 📝 **Account Renaming**: Customize account names during migration
- 🚀 **Batch Import**: Migrate all detected accounts with one command

**Example output:**

```
✅ Switched to account 'personal' globally
👤 Name: johndoe
📧 Email: john@example.com
🔗 SSH Host: github.com-personal

💡 To clone repositories with this account, use:
   git clone git@github.com-personal:username/repo.git
```

### SSH Configuration

When you generate a key, `krakncat` can automatically add an SSH config entry like this:

```ssh-config
Host github.com-personal
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_gh_personal
```

This allows you to use different SSH keys for different accounts by using the host alias:

```bash
git clone git@github.com-personal:username/repo.git
```

### Commands

| Command         | Description                                                               |
| --------------- | ------------------------------------------------------------------------- |
| `generate-key`  | Generate and configure a new SSH key for a GitHub account                 |
| `add`           | Add a new GitHub account with interactive prompts                         |
| `list` / `ls`   | List all configured accounts and current git configuration                |
| `use`           | Switch git configuration to use a specific account (globally or per-repo) |
| `config`        | Setup automatic git config for a directory using conditional includes     |
| `global`        | Set global git configuration to use a specific account                    |
| `show-includes` | Show current conditional includes in global git config                    |
| `migrate`       | Migrate existing git configuration to krakncat                            |
| `remove`        | Remove a GitHub account configuration                                     |
| `help`          | Show help for any command                                                 |

#### Key Flags

- `list --global` / `list -g`: Show only global git configuration
- `use [account] [path]`: Switch account globally or for specific repository

#### Flags for `generate-key`

- `--name` (required): Unique account name (e.g., 'work', 'personal')
- `--email` (required): Email address for the SSH key
- `--help`: Show help for the command

#### Arguments for `use`

- `account-name` (required): Name of the account to switch to
- `path` (optional): Repository path for local configuration (omit for global)

## Example Workflow

### Option 1: Manual Switching (Traditional)

1. **Generate SSH keys and add accounts:**

   ```bash
   # Add work account
   ./krakn generate-key --name work --email work@company.com
   # Follow prompts to save account configuration

   # Add personal account
   ./krakn add
   # Follow interactive prompts
   ```

2. **Switch between accounts as needed:**

   ```bash
   # Switch globally to work account
   ./krakn use work

   # Switch just for a specific project
   ./krakn use personal ~/my-personal-project
   ```

3. **Add the public keys to your GitHub accounts:**

   - Copy the displayed public key
   - Go to [GitHub SSH settings](https://github.com/settings/ssh/new)
   - Add the key with a descriptive title

4. **Clone repositories using the appropriate account:**

   ```bash
   # For work account
   git clone git@github.com-work:company/project.git

   # For personal account
   git clone git@github.com-personal:username/personal-project.git
   ```

### Option 2: Automatic Directory-Based Switching (Recommended! 🌟)

1. **Set up accounts and directory structure:**

   ```bash
   # Add accounts (same as above)
   ./krakn generate-key --name work --email work@company.com
   ./krakn add  # for personal

   # Set global default
   ./krakn global personal

   # Set up automatic switching for work directory
   ./krakn config ~/work work
   ```

2. **Now it's completely automatic:**

   ```bash
   cd ~/work/any-project        # Automatically uses work account
   git config user.email        # Shows: work@company.com

   cd ~/personal/my-project     # Automatically uses personal account
   git config user.email        # Shows: personal@gmail.com
   ```

### How It Works: Git Conditional Includes

Behind the scenes, `krakncat` modifies your `~/.gitconfig` to include:

```gitconfig
[user]
    name = Personal Name
    email = personal@gmail.com

[includeIf "gitdir:~/work/"]
    path = ~/work/.gitconfig
```

And creates `~/work/.gitconfig`:

```gitconfig
[user]
    name = Work Name
    email = work@company.com
```

This means:

- **Default**: Personal account everywhere
- **In ~/work/**: Work account automatically
- **No manual switching needed!** 🎉

## Account Storage

`krakncat` stores account configurations in `~/.krakncat/config.json`. This file contains:

- Account details (name, email, username, SSH key path)
- Current active account
- Account-specific settings

The configuration is automatically created when you add your first account.

## Project Structure

```
krakncat/
├── main.go              # Entry point
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── README.md            # This file
└── cmd/
    ├── root.go          # Root command definition
    ├── krakn.go         # generate-key command implementation
    ├── add.go           # add command implementation
    ├── list.go          # list command implementation
    ├── use.go           # use command implementation
    ├── directory.go     # config command implementation
    ├── global.go        # global and show-includes commands
    ├── migrate.go       # migrate command implementation
    ├── remove.go        # remove command implementation
    └── config.go        # Configuration management
```

## Upcoming Features

🚧 The following features are planned for future releases:

- `clone` - Clone repositories using the correct SSH key automatically
- `edit` - Edit account details
- `backup` - Backup/restore account configurations
- `clean` - Remove orphaned conditional includes from .gitconfig

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the [MIT License](LICENSE).

## Troubleshooting

### Build Issues

**Error: "found packages cmd and main"**

- This occurs when there are duplicate files. Ensure there's no `krakn.go` in the root directory.
- Solution: Remove any duplicate files and rebuild.

**Error: "go: command not found"**

- Go is not installed on your system.
- **Arch Linux**: `sudo pacman -S go`
- **Other distros**: Follow the [official Go installation guide](https://golang.org/doc/install)

**Error: Module download issues**

- Check your internet connection and Go proxy settings.
- Try: `go env -w GOPROXY=direct`

### Runtime Issues

**Error: "ssh-keygen: command not found"**

- OpenSSH is not installed.
- **Arch Linux**: `sudo pacman -S openssh`

**Permission denied accessing SSH files**

- Ensure proper permissions on `~/.ssh/` directory: `chmod 700 ~/.ssh`
- SSH key files should have `600` permissions: `chmod 600 ~/.ssh/id_*`

**Git config not working**

- Check if Git is installed: `git --version`
- Verify config with: `git config --list`

### Getting Help

- Run `krakn --help` for command overview
- Run `krakn <command> --help` for specific command help
- Check existing SSH configs: `cat ~/.ssh/config`
- View current Git config: `git config --global --list`
