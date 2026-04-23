# Troubleshooting Guide

This guide helps you resolve common issues with KDE Dotfiles Manager.

## Installation Issues

### Go Version Too Old

**Error:**
```
error: package requires Go 1.22 or higher
found: go1.19
```

**Solution:**
1. Download latest Go from [golang.org](https://golang.org/dl/)
2. Or use your package manager:
   ```bash
   # Ubuntu/Debian
   sudo apt update && sudo apt install golang-go
   
   # Fedora
   sudo dnf install golang
   
   # Arch Linux
   sudo pacman -S go
   ```
3. Verify installation:
   ```bash
   go version
   ```

### KDE Plasma Version Not Detected

**Warning:**
```
warning: KDE Plasma 6+ not detected
```

**Solution:**
1. Check your Plasma version:
   ```bash
   plasmashell --version
   ```
2. If using Plasma 5, some features may not work correctly
3. Consider upgrading to Plasma 6 for full compatibility

### Build Fails

**Error:**
```
make: *** [Makefile:10: build] Error 1
```

**Solution:**
1. Ensure all dependencies are installed:
   ```bash
   go mod download
   ```
2. Clean and rebuild:
   ```bash
   make clean
   make build
   ```
3. Check Go environment:
   ```bash
   go env
   ```

## Backup Issues

### Backup Fails with Permission Denied

**Error:**
```
Error: permission denied: ~/.config/kglobalshortcutsrc
```

**Solution:**
1. Check file permissions:
   ```bash
   ls -la ~/.config/kglobalshortcutsrc
   ```
2. Fix permissions if needed:
   ```bash
   chmod 644 ~/.config/kglobalshortcutsrc
   ```
3. Ensure you own the files:
   ```bash
   chown $USER:$USER ~/.config/*
   ```

### Backup Directory Not Created

**Error:**
```
Error: cannot create backup directory
```

**Solution:**
1. Manually create the directory:
   ```bash
   mkdir -p ~/kde-dotfiles
   ```
2. Check disk space:
   ```bash
   df -h ~
   ```
3. Verify write permissions:
   ```bash
   touch ~/kde-dotfiles/test && rm ~/kde-dotfiles/test
   ```

### Incomplete Backup

**Problem:** Some configuration files are missing from backup

**Solution:**
1. Run backup with verbose output:
   ```bash
   ./scripts/backup.sh --verbose
   ```
2. Check which files are being backed up:
   ```bash
   find ~/kde-dotfiles -type f
   ```
3. Verify source files exist:
   ```bash
   ls -la ~/.config/plasma-org.kde.plasma.desktop-appletsrc
   ```

## Restore Issues

### Restore Fails

**Error:**
```
Error: restore failed for category widgets
```

**Solution:**
1. Check backup integrity:
   ```bash
   ls -la ~/kde-dotfiles/widgets/
   ```
2. Try dry-run first:
   ```bash
   ./scripts/restore.sh --dry-run
   ```
3. Restore categories individually:
   ```bash
   ./scripts/restore.sh --category shortcuts
   ./scripts/restore.sh --category themes
   ```

### Configuration Not Applied After Restore

**Problem:** Settings appear unchanged after restore

**Solution:**
1. Restart Plasma shell:
   ```bash
   kquitapp6 plasmashell && plasmashell &
   ```
2. Log out and log back in
3. Verify files were copied:
   ```bash
   diff ~/kde-dotfiles/shortcuts/kglobalshortcutsrc ~/.config/kglobalshortcutsrc
   ```

### Widgets Not Restoring

**Problem:** Widgets don't appear after restore

**Solution:**
1. Check if widgets are installed:
   ```bash
   plasmapkg2 -l | grep widget-name
   ```
2. Install missing widgets manually
3. Re-run restore with widget category only:
   ```bash
   ./scripts/restore.sh --category widgets
   ```
4. Reset Plasma layout:
   ```bash
   mv ~/.config/plasma-org.kde.plasma.desktop-appletsrc{,.bak}
   kquitapp6 plasmashell && plasmashell &
   ```

## Widget Issues

### plasmapkg2 Not Found

**Error:**
```
warning: plasmapkg2 not found, widget installation disabled
```

**Solution:**
1. Install KDE Plasma workspace:
   ```bash
   # Ubuntu/Debian
   sudo apt install plasma-workspace
   
   # Fedora
   sudo dnf install plasma-workspace
   
   # Arch Linux
   sudo pacman -S plasma-workspace
   ```
2. Alternative: Manual widget installation:
   ```bash
   cp -r widget-folder ~/.local/share/plasma/plasmoids/
   ```

### Widget Installation Fails

**Error:**
```
Error: failed to install widget com.github.widget
```

**Solution:**
1. Check widget package integrity:
   ```bash
   ls -la ~/kde-dotfiles/widgets/plasmoids/
   ```
2. Try manual installation:
   ```bash
   cp -r ~/kde-dotfiles/widgets/plasmoids/widget-name ~/.local/share/plasma/plasmoids/
   ```
3. Verify permissions:
   ```bash
   chmod -R 755 ~/.local/share/plasma/plasmoids/widget-name
   ```
4. Check for metadata.json:
   ```bash
   cat ~/.local/share/plasma/plasmoids/widget-name/metadata.json
   ```

### Custom Widgets Not Detected

**Problem:** Custom widgets not identified during restore

**Solution:**
1. Verify widget ID format (should not start with org.kde.)
2. Check appletsrc file for widget entries:
   ```bash
   grep "widget-id" ~/kde-dotfiles/widgets/plasma-org.kde.plasma.desktop-appletsrc
   ```
3. Ensure widget is in backup manifest

## Sync Issues

### Git Authentication Fails

**Error:**
```
Error: git authentication failed
```

**Solution:**
1. Use SSH keys instead of HTTPS:
   ```bash
   git remote set-url origin git@github.com:user/repo.git
   ```
2. Generate SSH key if needed:
   ```bash
   ssh-keygen -t ed25519 -C "your_email@example.com"
   ```
3. Add key to GitHub SSH settings
4. Test connection:
   ```bash
   ssh -T git@github.com
   ```

### Sync Conflicts

**Error:**
```
Error: merge conflicts detected
```

**Solution:**
1. Navigate to dotfiles directory:
   ```bash
   cd ~/kde-dotfiles
   ```
2. Check conflict status:
   ```bash
   git status
   ```
3. Resolve conflicts manually
4. Commit resolution:
   ```bash
   git add .
   git commit -m "Resolve merge conflicts"
   git push
   ```

### Remote Repository Not Accessible

**Error:**
```
Error: cannot access remote repository
```

**Solution:**
1. Verify repository URL:
   ```bash
   git remote -v
   ```
2. Check network connectivity
3. Verify repository exists and is accessible
4. Update repository URL if needed:
   ```bash
   ./scripts/sync.sh --init git@github.com:user/new-repo.git
   ```

## TUI Issues

### TUI Doesn't Start

**Error:**
```
Error: terminal too small
```

**Solution:**
1. Resize terminal window
2. Minimum size is 80x24
3. Try fullscreen mode

### TUI Display Issues

**Problem:** Garbled characters or display corruption

**Solution:**
1. Ensure terminal supports UTF-8
2. Try different terminal emulator
3. Set proper locale:
   ```bash
   export LANG=en_US.UTF-8
   export LC_ALL=en_US.UTF-8
   ```
4. Clear terminal and restart:
   ```bash
   reset
   ./bin/kdm
   ```

### Keyboard Input Not Working

**Problem:** Arrow keys or other inputs don't work

**Solution:**
1. Check terminal mode
2. Try alternative keys (j/k for navigation)
3. Restart TUI
4. Check for conflicting terminal shortcuts

## Performance Issues

### Slow Backup/Restore

**Problem:** Operations take too long

**Solution:**
1. Exclude large unnecessary files
2. Enable compression:
   ```yaml
   backup:
     compression: true
   ```
3. Use selective category backup:
   ```bash
   ./scripts/backup.sh --category shortcuts,themes
   ```
4. Check disk I/O performance

### High Memory Usage

**Problem:** Application uses excessive memory

**Solution:**
1. Reduce number of categories processed simultaneously
2. Process categories sequentially
3. Close other applications
4. Check for memory leaks (report if found)

## Configuration Issues

### Config File Not Loaded

**Problem:** Default settings used instead of config file

**Solution:**
1. Verify file location:
   ```bash
   ls -la ~/.config/kde-dotfiles-manager/config.yaml
   ```
2. Check YAML syntax:
   ```bash
   python3 -c "import yaml; yaml.safe_load(open('~/.config/kde-dotfiles-manager/config.yaml'))"
   ```
3. Validate configuration:
   ```bash
   ./bin/kdm --validate-config
   ```

### Invalid Configuration Value

**Error:**
```
Error: invalid value for 'categories': unknown category 'invalid'
```

**Solution:**
1. Check available categories in documentation
2. Valid categories: shortcuts, themes, window_management, languages, widgets, panels
3. Fix configuration file
4. Use default config as reference

## Common Error Messages

### "No such file or directory"

**Cause:** Missing configuration files or directories

**Solution:**
1. Verify paths in error message
2. Create missing directories
3. Check for typos in paths

### "Permission denied"

**Cause:** Insufficient file permissions

**Solution:**
1. Check file ownership
2. Fix permissions with chmod
3. Run as appropriate user

### "Command not found"

**Cause:** Missing dependencies

**Solution:**
1. Install required packages
2. Check PATH environment variable
3. Verify script shebang lines

## Getting Help

### Debug Mode

Enable debug output for troubleshooting:

```bash
./bin/kdm --verbose --debug
./scripts/backup.sh --verbose --debug
```

### Log Files

Check log files for errors:

```bash
# Application logs
~/.local/share/kde-dotfiles-manager/logs/

# System logs
journalctl -u kde-dotfiles-manager
```

### Reporting Issues

When reporting issues, include:

1. KDE Plasma version: `plasmashell --version`
2. Go version: `go version`
3. Distribution and version
4. Error messages (full output)
5. Steps to reproduce
6. Configuration file (sanitized)

### Community Resources

- GitHub Issues: https://github.com/azyoskol/kde-dotfiles-manager/issues
- KDE Forums: https://forum.kde.org/
- Reddit: r/KDE

## Preventive Measures

### Regular Maintenance

1. **Update regularly**: Keep application updated
2. **Test backups**: Periodically verify backups work
3. **Monitor disk space**: Ensure adequate space for backups
4. **Document changes**: Keep track of configuration changes

### Best Practices

1. Always create backups before major changes
2. Test restores on non-production systems first
3. Keep multiple backup versions
4. Use version control for configurations
5. Document custom configurations

## See Also

- [Installation Guide](installation.md) - Setup instructions
- [Usage Guide](usage.md) - How to use the application
- [Configuration](configuration.md) - Configuration reference
- [Widget Management](widgets.md) - Widget-specific help
