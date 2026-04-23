# Usage Guide

This guide explains how to use KDE Dotfiles Manager for backing up, restoring, and managing your KDE Plasma 6+ configurations.

## Quick Start

### Launch the TUI Application

```bash
./bin/kdm
```

Or if installed system-wide:

```bash
kdm
```

### Using Bash Scripts Directly

```bash
# Backup all configurations
./scripts/backup.sh

# Restore from backup
./scripts/restore.sh

# Sync with remote repository
./scripts/sync.sh

# Deploy configuration
./scripts/deploy.sh
```

## TUI Interface

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Confirm selection |
| `Esc` / `q` | Go back / Quit |
| `Space` | Toggle checkbox |
| `r` | Refresh status |
| `d` | Deploy selected profile |
| `?` | Show help |

### Main Menu

The main menu provides access to all major functions:

```
┌─────────────────────────────────────────┐
│  KDE Dotfiles Manager                   │
├─────────────────────────────────────────┤
│  > Backup Configuration                 │
│    Restore Configuration                │
│    Synchronize                          │
│    Deploy Profile                       │
│    Settings                             │
│    Exit                                 │
└─────────────────────────────────────────┘
```

All operations are performed through the TUI interface - no external scripts required. The Go application handles backup, restore, sync, and deployment natively.

## Backup Configuration

### Creating a Full Backup

1. Launch the TUI: `./bin/kdm`
2. Select "Backup Configuration"
3. Choose categories to backup (or select all)
4. Press Enter to start the backup
5. Wait for completion

### Backup Categories

| Category | Files Backed Up |
|----------|----------------|
| **Shortcuts** | `kglobalshortcutsrc`, `khotkeysrc` |
| **Themes** | Color schemes, window decorations, cursors, icons, wallpapers, GTK themes |
| **Window Management** | KWin rules, virtual desktops, tiling settings, kwinrc |
| **Languages** | Locale settings, keyboard layouts, input methods |
| **Widgets** | Desktop widgets, panel applets, configurations with auto-install support |
| **Panels** | Panel layouts, positions, settings |
| **System Settings** | General system settings, power management, screen locker |

### Backup Structure

Backups are organized as follows:

```
~/kde-dotfiles/
├── shortcuts/
│   ├── kglobalshortcutsrc
│   └── khotkeysrc
├── themes/
│   ├── kdeglobals
│   ├── plasmarc
│   ├── color-schemes/
│   ├── wallpapers/
│   └── icons/
├── window_management/
│   ├── kwinrc
│   └── kwinrulesrc
├── languages/
│   ├── plasma-localerc
│   └── language.conf
├── widgets/
│   ├── plasma/
│   │   ├── plasmoids/
│   │   │   └── [widget-id]/
│   │   └── org.kde.plasma.desktop-appletsrc
│   └── org.kde.panel-appletsrc
├── panels/
│   └── org.kde.panel
└── system_settings/
    ├── kdeglobals
    ├── powerdevilrc
    └── kscreenlockerrc
```

## Restore Configuration

### Restoring from Backup

1. Launch the TUI: `./bin/kdm`
2. Select "Restore Configuration"
3. Choose the backup profile to restore
4. The system will automatically restore all available categories
5. If custom widgets are found, you'll be prompted to install them
6. Wait for completion and restart Plasma if needed

### Widget Auto-Installation

When restoring widget configurations:

1. The system scans for custom widgets in the backup
2. Checks if widgets are installed on the current system
3. Prompts to install missing widgets through the TUI
4. Installs widgets using the built-in Go installer (no external scripts)
5. Continues with configuration restore

### Safety Features

- **Automatic Pre-Restore Backup**: Creates a backup before restoring (configurable)
- **Profile Isolation**: Each profile has its own backup directory
- **Category Detection**: Only restores categories that exist in the backup

## Synchronization

### Setting Up Sync

1. Configure Git repository in settings or config file
2. Initialize sync with `./scripts/sync.sh --init`
3. Push initial backup: `./scripts/sync.sh --push`

### Sync Operations

```bash
# Initialize sync with remote repository
./scripts/sync.sh --init git@github.com:user/kde-dotfiles.git

# Push local changes to remote
./scripts/sync.sh --push

# Pull changes from remote
./scripts/sync.sh --pull

# Sync both ways (pull then push)
./scripts/sync.sh --sync

# Check sync status
./scripts/sync.sh --status
```

### Multi-Machine Setup

1. Set up backup on Machine A
2. Configure Git repository
3. Push backup: `./scripts/sync.sh --push`
4. On Machine B, pull and restore:
   ```bash
   ./scripts/sync.sh --pull
   ./scripts/restore.sh
   ```

## Deployment

### Deploy Profiles

Profiles allow you to maintain different configurations:

```bash
# List available profiles
./scripts/deploy.sh --list

# Deploy specific profile
./scripts/deploy.sh --profile default

# Deploy with category selection
./scripts/deploy.sh --profile work --category shortcuts,themes
```

### Profile Management

Create different profiles for different use cases:

- **default**: Standard desktop setup
- **work**: Work-specific shortcuts and layout
- **gaming**: Gaming-optimized settings
- **minimal**: Minimalist configuration

## Configuration File

Location: `~/.config/kde-dotfiles-manager/config.yaml`

```yaml
# Directory to store dotfiles backups
dotfiles_dir: "~/kde-dotfiles"

# Git repository URL for synchronization
git_repo: "git@github.com:user/kde-dotfiles.git"

# Categories to include in backup
categories:
  - shortcuts
  - themes
  - window_management
  - languages
  - widgets
  - panels

# Auto-create backup before restore
backup_before_restore: true

# Verbose logging
verbose: false

# Widget installation settings
widgets:
  auto_install: true
  install_path: "~/.local/share/plasma/plasmoids"
```

## Advanced Usage

### Selective Backup

Backup only specific configuration files:

```bash
./scripts/backup.sh --files kglobalshortcutsrc,kwinrc
```

### Custom Backup Location

```bash
./scripts/backup.sh --output /path/to/backup
```

### Restore to Different Location

For testing configurations:

```bash
./scripts/restore.sh --target /tmp/test-kde-config
```

### Compare Configurations

Compare current config with backup:

```bash
diff -u ~/.config/kglobalshortcutsrc ~/kde-dotfiles/shortcuts/kglobalshortcutsrc
```

## Best Practices

### Regular Backups

- Create backups after significant configuration changes
- Schedule regular backups (weekly/monthly)
- Keep multiple backup versions

### Version Control

- Use Git for tracking configuration changes
- Commit with descriptive messages
- Tag important configurations

### Testing Restores

- Periodically test restore process
- Verify widget installations work correctly
- Check shortcut conflicts

### Documentation

- Document custom configurations
- Note widget dependencies
- Keep track of theme sources

## Troubleshooting

### Common Issues

#### Backup Fails

**Problem:** Permission denied errors

**Solution:** Ensure read access to `~/.config/` directory

#### Restore Fails

**Problem:** Configuration not applied

**Solution:** Restart Plasma shell:
```bash
kquitapp5 plasmashell && plasmashell &
```

For Plasma 6:
```bash
kquitapp6 plasmashell && plasmashell &
```

#### Widgets Not Installing

**Problem:** plasmapkg command not found

**Solution:** Install KDE Plasma workspace packages or manually copy widgets:
```bash
cp -r backup/widgets/plasmoids/* ~/.local/share/plasma/plasmoids/
```

#### Sync Conflicts

**Problem:** Git merge conflicts

**Solution:** Manually resolve conflicts in the dotfiles directory, then commit

## Next Steps

- Explore [Widget Management](widgets.md) for detailed widget handling
- Check [Troubleshooting](troubleshooting.md) for common issues
- Review [Configuration](configuration.md) for all options
