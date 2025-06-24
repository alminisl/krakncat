#!/bin/bash

# krakncat installation script for Arch Linux
# This script will install dependencies and build krakncat

set -e

echo "🐙 krakncat Installation Script for Arch Linux"
echo "==============================================="

# Check if running on Arch Linux
if [ ! -f /etc/arch-release ]; then
    echo "⚠️  This script is designed for Arch Linux."
    echo "   You can still run it, but package installation may fail."
    read -p "   Continue anyway? [y/N]: " continue_anyway
    if [[ ! $continue_anyway =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "📦 Installing Go..."
    sudo pacman -S --needed go
else
    echo "✅ Go is already installed: $(go version)"
fi

# Check if git is installed
if ! command -v git &> /dev/null; then
    echo "📦 Installing Git..."
    sudo pacman -S --needed git
else
    echo "✅ Git is already installed: $(git --version)"
fi

# Check if ssh-keygen is available
if ! command -v ssh-keygen &> /dev/null; then
    echo "📦 Installing OpenSSH..."
    sudo pacman -S --needed openssh
else
    echo "✅ SSH tools are available"
fi

echo ""
echo "🔨 Building krakncat..."

# Ensure we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Please run this script from the krakncat directory."
    exit 1
fi

# Build the application
go mod tidy
go build -o krakn .

if [ $? -eq 0 ]; then
    echo "✅ krakncat built successfully!"
    
    # Ask if user wants to install system-wide
    read -p "📍 Install krakncat to /usr/local/bin? [y/N]: " install_system
    if [[ $install_system =~ ^[Yy]$ ]]; then
        sudo cp krakn /usr/local/bin/
        echo "✅ krakncat installed to /usr/local/bin/krakn"
        echo ""
        echo "🎉 Installation complete! You can now run 'krakn' from anywhere."
    else
        echo "✅ krakncat is ready! Run './krakn' to start."
    fi
    
    echo ""
    echo "🚀 Quick start:"
    echo "   krakn --help          # Show all commands"
    echo "   krakn list            # List accounts (triggers migration if first run)"
    echo "   krakn add             # Add a new GitHub account"
    echo ""
else
    echo "❌ Build failed! Please check the error messages above."
    exit 1
fi
