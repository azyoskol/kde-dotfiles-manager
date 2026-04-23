# Configuration Reference

This document provides a comprehensive reference for configuring KDE Dotfiles Manager.

## Configuration File Location

The main configuration file is located at:

```
~/.config/kde-dotfiles-manager/config.yaml
```

## Complete Configuration Example

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
  skip_kde_widgets: true

# Sync settings
sync:
  auto_push: false
  auto_pull: false
  branch: "main"

# Backup settings
backup:
  compression: true
  timestamp_format: "2006-01-02_15-04-05"
  max_backups: 10

# Restore settings
restore:
  confirm_overwrite: true
  dry_run: false
  backup_current: true
```

## Configuration Options

### Core Settings

#### `dotfiles_dir` (string)

**Default:** `"~/kde-dotfiles"`

Directory where all backups and configurations are stored.

```yaml
dotfiles_dir: "~/kde-dotfiles"
# or
dotfiles_dir: "/path/to/custom/location"
```

#### `git_repo` (string, optional)

**Default:** `null`

Git repository URL for synchronization. Required for sync functionality.

```yaml
git_repo: "git@github.com:user/kde-dotfiles.git"
# or
git_repo: "https://github.com/user/kde-dotfiles.git"
```

#### `categories` (list)

**Default:** All categories

List of configuration categories to include in backup.

Available categories:
- `shortcuts` - Keyboard shortcuts
- `themes` - Visual themes and appearance
- `window_management` - KWin rules and window behavior
- `languages` - Locale and language settings
- `widgets` - Desktop widgets and plasmoids
- `panels` - Panel configurations

```yaml
categories:
  - shortcuts
  - themes
  - widgets
```

### Backup Settings

#### `backup_before_restore` (boolean)

**Default:** `true`

Automatically create a backup before restoring configurations.

```yaml
backup_before_restore: true
```

#### `verbose` (boolean)

**Default:** `false`

Enable verbose logging for debugging.

```yaml
verbose: true
```

#### `backup.compression` (boolean)

**Default:** `true`

Compress backup files to save space.

```yaml
backup:
  compression: true
```

#### `backup.timestamp_format` (string)

**Default:** `"2006-01-02_15-04-05"`

Go time format for backup timestamps.

```yaml
backup:
  timestamp_format: "2006-01-02_15-04-05"
```

#### `backup.max_backups` (integer)

**Default:** `10`

Maximum number of backups to keep. Older backups are automatically deleted.

```yaml
backup:
  max_backups: 10
```

### Widget Settings

#### `widgets.auto_install` (boolean)

**Default:** `true`

Automatically install missing widgets during restore.

```yaml
widgets:
  auto_install: true
```

#### `widgets.install_path` (string)

**Default:** `"~/.local/share/plasma/plasmoids"`

Directory where widgets are installed.

```yaml
widgets:
  install_path: "~/.local/share/plasma/plasmoids"
```

#### `widgets.skip_kde_widgets` (boolean)

**Default:** `true`

Skip installation of official KDE widgets (org.kde.*).

```yaml
widgets:
  skip_kde_widgets: true
```

### Sync Settings

#### `sync.auto_push` (boolean)

**Default:** `false`

Automatically push changes after backup.

```yaml
sync:
  auto_push: true
```

#### `sync.auto_pull` (boolean)

**Default:** `false`

Automatically pull changes before restore.

```yaml
sync:
  auto_pull: true
```

#### `sync.branch` (string)

**Default:** `"main"`

Git branch to use for synchronization.

```yaml
sync:
  branch: "main"
```

### Restore Settings

#### `restore.confirm_overwrite` (boolean)

**Default:** `true`

Ask for confirmation before overwriting existing configurations.

```yaml
restore:
  confirm_overwrite: true
```

#### `restore.dry_run` (boolean)

**Default:** `false`

Show what would be restored without actually restoring.

```yaml
restore:
  dry_run: false
```

#### `restore.backup_current` (boolean)

**Default:** `true`

Backup current configuration before restoring.

```yaml
restore:
  backup_current: true
```

## Environment Variables

Configuration can also be overridden using environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `KDM_DOTFILES_DIR` | Override dotfiles directory | `export KDM_DOTFILES_DIR=~/my-backups` |
| `KDM_GIT_REPO` | Override Git repository | `export KDM_GIT_REPO=git@github.com:user/repo.git` |
| `KDM_VERBOSE` | Enable verbose mode | `export KDM_VERBOSE=1` |
| `KDM_CONFIG` | Custom config file path | `export KDM_CONFIG=~/.custom-config.yaml` |

## Command Line Overrides

Most configuration options can be overridden via command line flags:

```bash
# Override dotfiles directory
./bin/kdm --dotfiles-dir ~/custom-backups

# Enable verbose mode
./bin/kdm --verbose

# Specify categories
./scripts/backup.sh --category shortcuts,themes

# Custom config file
./bin/kdm --config /path/to/config.yaml
```

## Profile-Specific Configuration

Create multiple configuration profiles for different scenarios:

### Work Profile

```yaml
# ~/.config/kde-dotfiles-manager/work.yaml
dotfiles_dir: "~/kde-dotfiles-work"
categories:
  - shortcuts
  - window_management
widgets:
  auto_install: false
```

### Gaming Profile

```yaml
# ~/.config/kde-dotfiles-manager/gaming.yaml
dotfiles_dir: "~/kde-dotfiles-gaming"
categories:
  - shortcuts
  - themes
  - window_management
```

Use profiles with:

```bash
./bin/kdm --profile work
./scripts/backup.sh --config ~/.config/kde-dotfiles-manager/gaming.yaml
```

## Configuration Validation

The application validates configuration on startup. Common validation errors:

### Invalid Path

```
Error: dotfiles_dir "~/invalid~path" is not accessible
```

**Solution:** Ensure the directory exists and is writable.

### Invalid Category

```
Error: unknown category "invalid_category"
```

**Solution:** Use only valid category names from the list above.

### Invalid Git Repository

```
Error: git_repo "invalid-url" is not a valid Git repository
```

**Solution:** Verify the repository URL and your access permissions.

## Migration from Previous Versions

### From v1.x to v2.x

Configuration format changed in v2.0. To migrate:

1. Backup old configuration
2. Update format:
   ```yaml
   # Old format
   backup_dir: ~/kde-dotfiles
   
   # New format
   dotfiles_dir: ~/kde-dotfiles
   ```
3. Run migration tool:
   ```bash
   ./bin/kdm --migrate-config
   ```

## Troubleshooting Configuration

### Configuration Not Loading

**Problem:** Application uses default settings

**Solution:** 
1. Check file location: `ls -la ~/.config/kde-dotfiles-manager/config.yaml`
2. Verify YAML syntax: `python3 -c "import yaml; yaml.safe_load(open('~/.config/kde-dotfiles-manager/config.yaml'))"`
3. Check permissions: `chmod 644 ~/.config/kde-dotfiles-manager/config.yaml`

### Settings Not Applied

**Problem:** Configuration changes have no effect

**Solution:**
1. Restart the application
2. Check for syntax errors in YAML
3. Verify option names match documentation
4. Check environment variable overrides

### Widget Installation Fails

**Problem:** Widgets don't install automatically

**Solution:**
1. Verify `widgets.auto_install: true`
2. Check `widgets.install_path` exists
3. Ensure `plasmapkg2` or `plasmapkg` is available
4. Set `widgets.skip_kde_widgets: false` if installing KDE widgets

## Best Practices

1. **Version Control Your Config**: Keep your configuration file in version control
2. **Document Custom Settings**: Comment non-default configurations
3. **Test Changes**: Use `--dry-run` to test configuration changes
4. **Backup Regularly**: Maintain multiple backup versions
5. **Use Profiles**: Separate configs for different use cases

## See Also

- [Installation Guide](installation.md) - Setup instructions
- [Usage Guide](usage.md) - How to use the application
- [Widget Management](widgets.md) - Widget-specific configuration
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
