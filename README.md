# KDE Dotfiles Manager

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org/)
[![KDE Plasma](https://img.shields.io/badge/KDE-Plasma_6+-brightgreen.svg)](https://kde.org/plasma-desktop/)

A powerful Terminal User Interface (TUI) application for managing KDE Plasma 6+ configuration backups, synchronization, and deployment with automatic widget restoration.

## ✨ Features

- **📦 Backup & Restore**: Save and restore complete KDE Plasma 6+ configurations
- **🔄 Synchronization**: Sync configurations across multiple machines via Git
- **🚀 Deployment**: Quick deployment of saved configurations
- **🎨 Beautiful TUI**: Interactive terminal interface built with Bubbletea
- **🧩 Widget Auto-Install**: Automatically detect and install custom widgets during restore
- **🛡️ Safe Operations**: Automatic backup creation before restore operations

### 📋 Managed Configuration Categories

| Category | Description |
|----------|-------------|
| **Window Management** | KWin rules, window behavior, virtual desktops, tiling scripts |
| **Keyboard Shortcuts** | Global shortcuts, KWin shortcuts, application shortcuts |
| **Themes** | Color schemes, window decorations, cursors, icons, wallpapers, GTK themes |
| **Languages** | System locale, input methods, spell checking, keyboard layouts |
| **Widgets** | Desktop widgets, panel configurations, desktop layout with auto-install |
| **Panels** | Panel layouts, applets, and configurations |

## 📖 Documentation

For detailed documentation, see the [docs/](docs/) directory:

- [Installation Guide](docs/installation.md) - Complete installation instructions
- [Usage Guide](docs/usage.md) - How to use the TUI and CLI tools
- [Configuration](docs/configuration.md) - Configuration file reference
- [Widget Management](docs/widgets.md) - Widget backup and auto-installation
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions

## ⚙️ Prerequisites

- Linux with KDE Plasma 6+
- Go 1.22+
- Bash 5.0+
- Git (for sync functionality)
- `plasmapkg2` or `plasmapkg` (for widget management)

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/azyoskol/kde-dotfiles-manager.git
cd kde-dotfiles-manager

# Build the TUI application
make build

# Launch the application
./bin/kdm
```

## 📸 Screenshots

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

## 🔧 Basic Commands

```bash
# Launch TUI application
./bin/kdm

# The application provides a complete TUI for all operations
# No external scripts required - everything is built into the Go binary
```

## 📁 Project Structure

```
kde-dotfiles-manager/
├── cmd/kdm/main.go          # TUI application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── kde/                 # KDE-specific path definitions
│   ├── sync/                # Git synchronization logic
│   ├── theme/               # Theme configuration handling
│   ├── widgets/             # Widget management & auto-install
│   ├── shortcuts/           # Keyboard shortcut handling
│   ├── locales/             # Language/locale configuration
│   └── backup/              # Backup and restore operations (pure Go)
├── docs/                    # Documentation
├── assets/                  # Static assets
├── Makefile                 # Build automation
└── README.md
```

## ⌨️ TUI Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` | Select/Confirm |
| `Esc` / `q` | Go back / Quit |
| `Space` | Toggle selection |
| `r` | Refresh status |
| `d` | Deploy selected profile |

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

**Made with ❤️ for the KDE Community**
