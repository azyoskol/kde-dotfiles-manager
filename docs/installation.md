# Installation Guide

This guide provides detailed instructions for installing KDE Dotfiles Manager on your system.

## Prerequisites

Before installing KDE Dotfiles Manager, ensure you have the following:

### Required Software

| Software | Version | Purpose |
|----------|---------|---------|
| KDE Plasma | 6.0+ | Desktop environment |
| Go | 1.22+ | Build toolchain |
| Bash | 5.0+ | Script execution |
| Git | Any | Version control & sync |

### Optional Dependencies

| Software | Purpose |
|----------|---------|
| `plasmapkg2` or `plasmapkg` | Widget installation |
| `rsync` | Fast file synchronization |
| `diff` | Configuration comparison |

## Installation Methods

### Method 1: From Source (Recommended)

#### Step 1: Clone the Repository

```bash
git clone https://github.com/azyoskol/kde-dotfiles-manager.git
cd kde-dotfiles-manager
```

#### Step 2: Verify Go Installation

```bash
go version
# Should output: go version go1.22.x ...
```

If Go is not installed, download it from [golang.org](https://golang.org/dl/) or use your package manager:

```bash
# Ubuntu/Debian
sudo apt install golang-go

# Fedora
sudo dnf install golang

# Arch Linux
sudo pacman -S go
```

#### Step 3: Build the Application

```bash
make build
```

This compiles the TUI application and places the binary in `./bin/kdm`.

#### Step 4: Install Scripts (Optional)

```bash
make install-scripts
```

This installs the bash scripts to `/usr/local/bin/` for system-wide access.

#### Step 5: Verify Installation

```bash
./bin/kdm --version
```

### Method 2: Using Go Install

If you have Go configured with `$GOPATH/bin` in your `$PATH`:

```bash
go install github.com/azyoskol/kde-dotfiles-manager/cmd/kdm@latest
```

The binary will be available at `$GOPATH/bin/kdm`.

### Method 3: Manual Binary Installation

1. Download the latest release from the [Releases page](https://github.com/azyoskol/kde-dotfiles-manager/releases)
2. Extract the archive:
   ```bash
   tar -xzf kdm-linux-amd64.tar.gz
   ```
3. Move to your PATH:
   ```bash
   sudo mv kdm /usr/local/bin/
   ```

## Post-Installation Setup

### Create Configuration Directory

```bash
mkdir -p ~/.config/kde-dotfiles-manager
```

### Initialize Default Configuration

The application will create a default configuration on first run. You can also create it manually:

```bash
cat > ~/.config/kde-dotfiles-manager/config.yaml << EOF
dotfiles_dir: "~/kde-dotfiles"
categories:
  - shortcuts
  - themes
  - window_management
  - languages
  - widgets
  - panels
backup_before_restore: true
verbose: false
EOF
```

### Set Up Dotfiles Directory

```bash
mkdir -p ~/kde-dotfiles
```

This directory will store all your backed up configurations.

## Verifying KDE Plasma Version

Ensure you're running KDE Plasma 6+:

```bash
plasmashell --version
# Should output: plasmashell 6.x.x
```

Or check via system settings:
- Open System Settings
- Navigate to "About System"
- Check Plasma version

## Troubleshooting Installation

### Common Issues

#### Go Version Too Old

```
error: package requires Go 1.22 or higher
```

**Solution:** Update Go to the latest version from [golang.org](https://golang.org/dl/).

#### Missing KDE Plasma 6

```
warning: KDE Plasma 6+ not detected
```

**Solution:** Upgrade your KDE Plasma installation. The tool is designed specifically for Plasma 6+.

#### Permission Denied on Script Install

```
make: *** [install-scripts] Permission denied
```

**Solution:** Run with sudo or install to a user-writable directory:
```bash
make install-scripts PREFIX=~/.local
```

#### plasmapkg Not Found

```
warning: plasmapkg2/plasmapkg not found, widget installation disabled
```

**Solution:** Install KDE Plasma development packages:
```bash
# Ubuntu/Debian
sudo apt install plasma-workspace-dev

# Fedora
sudo dnf install plasma-workspace-devel

# Arch Linux
sudo pacman -S plasma-workspace
```

## Next Steps

After successful installation:

1. Read the [Usage Guide](usage.md) to learn how to use the application
2. Review the [Configuration](configuration.md) options
3. Try your first backup with `./bin/kdm` or `kdm` command

## Uninstallation

To remove KDE Dotfiles Manager:

```bash
# If installed via make
make uninstall

# Manual removal
sudo rm /usr/local/bin/kdm
sudo rm /usr/local/bin/kde-backup.sh
sudo rm /usr/local/bin/kde-restore.sh
sudo rm -rf ~/.config/kde-dotfiles-manager
```

Your backed up dotfiles in `~/kde-dotfiles` will remain intact.
