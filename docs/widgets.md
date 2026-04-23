# Widget Management Guide

This guide covers the widget backup, restore, and auto-installation features of KDE Dotfiles Manager.

## Overview

KDE Dotfiles Manager provides comprehensive widget management:

- **Backup**: Save widget configurations and custom plasmoids
- **Auto-Install**: Automatically install missing widgets during restore
- **Detection**: Identify custom vs. official KDE widgets
- **Restoration**: Restore widget positions and settings

## Widget Backup

### What Gets Backed Up

| Component | Location | Description |
|-----------|----------|-------------|
| **Widget Configs** | `~/.config/plasma-org.kde.plasma.desktop-appletsrc` | All widget settings and positions |
| **Panel Widgets** | `~/.config/org.kde.panel-appletsrc` | Panel-specific applets |
| **Custom Plasmoids** | `~/.local/share/plasma/plasmoids/` | Third-party widgets |
| **Widget Metadata** | Backup manifest | Widget IDs and versions |

### Running a Widget Backup

```bash
# Full backup including widgets
./scripts/backup.sh --category widgets

# Backup only widget configurations
./scripts/backup.sh --category widgets,panels

# Verbose widget backup
./scripts/backup.sh --category widgets --verbose
```

### Backup Structure

```
~/kde-dotfiles/widgets/
├── plasma-org.kde.plasma.desktop-appletsrc    # Desktop widget config
├── org.kde.panel-appletsrc                    # Panel widget config
└── plasmoids/                                 # Custom plasmoids
    ├── com.github.widget1/
    │   ├── contents/
    │   └── metadata.json
    └── com.example.widget2/
        ├── contents/
        └── metadata.json
```

## Widget Auto-Installation

### How It Works

When restoring configurations with widgets:

1. **Scan Backup**: System reads backed up widget configurations
2. **Extract Widget IDs**: Parses `plasma-org.kde.plasma.desktop-appletsrc` for widget identifiers
3. **Check Installation**: Verifies which widgets are installed on current system
4. **Filter KDE Widgets**: Skips official KDE widgets (org.kde.*) by default
5. **Prompt User**: Shows list of missing custom widgets
6. **Install Widgets**: Uses kpackagetool6 or kpackagetool5 to install
7. **Restore Config**: Applies widget configurations after installation

### Installation Methods

#### Method 1: kpackagetool6 (Plasma 6 - Recommended)

```bash
kpackagetool6 --install /path/to/widget
```

This is the primary method for KDE Plasma 6+. It can install widgets from:
- `.plasmoid` package files
- Unpacked widget directories (containing `metadata.json`)

#### Method 2: kpackagetool5 (Plasma 5 - Fallback)

```bash
kpackagetool5 --install /path/to/widget
```

Automatic fallback for older Plasma 5 installations.

#### Method 3: Direct Directory Installation

kpackagetool6 can install directly from unpacked widget directories:

```bash
kpackagetool6 --install ~/.local/share/plasma/plasmoids/widget-name
```

### Configuration

Enable/disable auto-installation in config:

```yaml
widgets:
  auto_install: true           # Enable auto-install
  install_path: "~/.local/share/plasma/plasmoids"
  skip_kde_widgets: true       # Skip org.kde.* widgets
```

## Restoring Widgets

### Full Restore with Widgets

```bash
# Restore all including widgets
./scripts/restore.sh --category widgets

# Interactive restore with confirmation
./scripts/restore.sh --category widgets --interactive

# Dry run to see what would be restored
./scripts/restore.sh --category widgets --dry-run
```

### Restore Process

1. **Pre-Restore Check**: Verify backup exists and is valid
2. **Widget Detection**: Identify widgets in backup
3. **Installation Prompt**: Show missing widgets
4. **Installation**: Install missing widgets
5. **Config Restoration**: Copy configuration files
6. **Plasma Restart**: Optionally restart Plasma shell

### TUI Widget Restore

In the TUI interface:

1. Select "Restore Configuration"
2. Choose backup
3. Select "Widgets" category
4. Review detected widgets
5. Confirm installation of missing widgets
6. Complete restoration

## Managing Custom Widgets

### Identifying Custom Widgets

Custom widgets are identified by their ID:

- **Official KDE**: `org.kde.*` (e.g., `org.kde.weather`)
- **Custom**: Any other namespace (e.g., `com.github.user.widget`)

### Listing Installed Widgets

```bash
# List all installed widgets (Plasma 6)
kpackagetool6 --list

# List all installed widgets (Plasma 5)
kpackagetool5 --list

# Or check directory directly
ls ~/.local/share/plasma/plasmoids/
```

### Finding Widget Sources

If a widget is missing:

1. Check the backup for `metadata.json`
2. Look for repository URL in metadata
3. Download from source (GitHub, KDE Store, etc.)
4. Install before or during restore

Example metadata.json:
```json
{
  "KPlugin": {
    "Id": "com.github.customwidget",
    "Name": "Custom Widget",
    "SourceUrl": "https://github.com/user/custom-widget"
  }
}
```

## Troubleshooting Widgets

### Widget Not Installing

**Problem:** Installation fails during restore

**Solutions:**
1. Check kpackagetool6 availability:
   ```bash
   which kpackagetool6
   ```
2. Verify widget package/directory integrity (must contain `metadata.json`)
3. Try manual installation:
   ```bash
   kpackagetool6 --install /path/to/widget
   ```
4. Check permissions:
   ```bash
   chmod -R 755 ~/.local/share/plasma/plasmoids/widget
   ```
5. Check for error messages:
   ```bash
   kpackagetool6 --install /path/to/widget 2>&1
   ```

### Widget Not Appearing

**Problem:** Widget installed but not showing on desktop

**Solutions:**
1. Restart Plasma shell:
   ```bash
   kquitapp6 plasmashell && plasmashell &
   ```
2. Check if widget is disabled:
   - Right-click desktop → Configure Desktop
   - Check widget visibility
3. Verify configuration file:
   ```bash
   cat ~/.config/plasma-org.kde.plasma.desktop-appletsrc
   ```

### Configuration Not Applied

**Problem:** Widget installed but settings not restored

**Solutions:**
1. Verify appletsrc file was copied
2. Check widget ID matches in config
3. Manually merge configurations
4. Remove and re-add widget to desktop

### Conflicting Widget Versions

**Problem:** Different widget version than backed up

**Solutions:**
1. Note version differences
2. Update widget to match backup version
3. Adjust configuration for new version
4. Consider version-locking important widgets

## Advanced Widget Management

### Exporting Specific Widgets

Export individual widget configurations:

```bash
# Extract specific widget config
grep -A 50 "\[com.github.widget\]" ~/.config/plasma-org.kde.plasma.desktop-appletsrc > widget-config.ini
```

### Importing Widget Configurations

Manually add widget to configuration:

```bash
# Append to appletsrc
cat widget-config.ini >> ~/.config/plasma-org.kde.plasma.desktop-appletsrc
```

### Widget Version Tracking

Track widget versions in your backup:

```bash
# Create version manifest
for widget in ~/.local/share/plasma/plasmoids/*/; do
  if [ -f "$widget/metadata.json" ]; then
    echo "$(basename $widget): $(jq -r '.KPlugin.Version' $widget/metadata.json)"
  fi
done > ~/kde-dotfiles/widgets/versions.txt
```

### Batch Widget Operations

Install multiple widgets:

```bash
for widget in ~/downloads/widgets/*; do
  kpackagetool6 --install "$widget"
done
```

Remove widgets:

```bash
for widget in com.github.widget1 com.github.widget2; do
  kpackagetool6 --uninstall "$widget"
done
```

## Best Practices

### Before Backup

1. **Update Widgets**: Ensure all widgets are updated
2. **Test Functionality**: Verify all widgets work correctly
3. **Note Dependencies**: Document external dependencies
4. **Clean Unused**: Remove unused widgets before backup

### During Restore

1. **Review List**: Check detected widgets before installing
2. **Verify Sources**: Ensure widgets are from trusted sources
3. **Test Incrementally**: Restore widgets in batches if many
4. **Keep Backups**: Maintain pre-restore backup

### Maintenance

1. **Regular Updates**: Keep widgets updated
2. **Version Control**: Track widget versions
3. **Document Changes**: Note when adding/removing widgets
4. **Test After Updates**: Verify widgets work after updates

## Widget Recommendations

### Popular Widget Categories

| Category | Examples |
|----------|----------|
| **System Monitoring** | CPU, RAM, Network usage |
| **Weather** | Various weather providers |
| **Productivity** | Notes, calendars, task lists |
| **Media Controls** | Player controls, volume |
| **News/RSS** | News feeds, RSS readers |
| **Custom Scripts** | User-created functionality |

### Finding Widgets

- **KDE Store**: https://store.kde.org/browse?category=plasma-applets
- **GitHub**: Search for "plasma widget" or "plasmoid"
- **KDE Apps**: https://apps.kde.org/

## Command Reference

### kpackagetool6 Commands (Plasma 6)

```bash
# Install widget (from .plasmoid or directory)
kpackagetool6 --install /path/to/widget

# Remove widget
kpackagetool6 --uninstall widget-id

# List installed widgets
kpackagetool6 --list

# Show widget info
kpackagetool6 --info widget-id

# Upgrade widget
kpackagetool6 --upgrade /path/to/widget
```

### kpackagetool5 Commands (Plasma 5 - Fallback)

```bash
# Install widget
kpackagetool5 --install /path/to/widget

# Remove widget
kpackagetool5 --uninstall widget-id

# List installed widgets
kpackagetool5 --list
```

### Backup Script Options

```bash
# Widget-specific backup
./scripts/backup.sh --category widgets

# Include custom plasmoids
./scripts/backup.sh --category widgets --include-plasmoids

# Exclude panel widgets
./scripts/backup.sh --category widgets --exclude-panels
```

### Restore Script Options

```bash
# Widget restore with auto-install
./scripts/restore.sh --category widgets --auto-install

# Skip widget installation
./scripts/restore.sh --category widgets --no-install

# Install only, skip config
./scripts/restore.sh --category widgets --install-only
```

## See Also

- [Usage Guide](usage.md) - General usage instructions
- [Configuration](configuration.md) - Widget configuration options
- [Troubleshooting](troubleshooting.md) - Common issues
