#!/usr/bin/env bash
# =============================================================================
# KDE Dotfiles Manager - Common Functions
# Shared utilities for all backup/restore/sync/deploy scripts
# =============================================================================

set -euo pipefail

# Color definitions
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly MAGENTA='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly BOLD='\033[1m'
readonly NC='\033[0m' # No Color

# Default configuration
DEFAULT_DOTFILES_DIR="${HOME}/kde-dotfiles"
DEFAULT_CATEGORIES="shortcuts,themes,window_management,languages,widgets,panels,system_settings"
VERBOSE=0

# =============================================================================
# Logging Functions
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_verbose() {
    if [[ "${VERBOSE}" -eq 1 ]]; then
        echo -e "${CYAN}[DEBUG]${NC} $*"
    fi
}

# =============================================================================
# Utility Functions
# =============================================================================

# Check if a command exists
command_exists() {
    command -v "$1" &>/dev/null
}

# Create directory if it doesn't exist
ensure_dir() {
    local dir="$1"
    if [[ ! -d "${dir}" ]]; then
        mkdir -p "${dir}"
        log_verbose "Created directory: ${dir}"
    fi
}

# Check if running on KDE Plasma
check_kde_plasma() {
    if ! command_exists plasmashell; then
        log_warning "plasmashell not found. Are you running KDE Plasma?"
        return 1
    fi

    local plasma_version
    plasma_version=$(plasmashell --version 2>/dev/null | grep -oP '\d+' | head -1)

    if [[ -z "${plasma_version}" ]] || [[ "${plasma_version}" -lt 6 ]]; then
        log_warning "KDE Plasma 6+ recommended. Detected version: ${plasma_version:-unknown}"
        return 1
    fi

    log_verbose "KDE Plasma ${plasma_version} detected"
    return 0
}

# Stop KDE services before modifying config files
stop_kde_services() {
    log_info "Stopping KDE services..."
    
    # Stop plasmashell
    if pgrep -x plasmashell &>/dev/null; then
        log_verbose "Stopping plasmashell..."
        killall plasmashell 2>/dev/null || true
        sleep 1
    fi

    # Stop kded6
    if pgrep -x kded6 &>/dev/null; then
        log_verbose "Stopping kded6..."
        killall kded6 2>/dev/null || true
        sleep 1
    fi
}

# Restart KDE services after modifying config files
start_kde_services() {
    log_info "Restarting KDE services..."
    
    # Restart kded6
    if ! pgrep -x kded6 &>/dev/null; then
        log_verbose "Starting kded6..."
        kded6 &>/dev/null &
        sleep 1
    fi

    # Restart plasmashell
    if ! pgrep -x plasmashell &>/dev/null; then
        log_verbose "Starting plasmashell..."
        plasmashell &>/dev/null &
        sleep 1
    fi
}

# Copy file with backup of existing destination
copy_with_backup() {
    local src="$1"
    local dest="$2"
    
    if [[ ! -f "${src}" ]]; then
        log_verbose "Source file not found: ${src}"
        return 1
    fi

    # Create parent directory if needed
    ensure_dir "$(dirname "${dest}")"

    # Backup existing file
    if [[ -f "${dest}" ]]; then
        local backup="${dest}.bak.$(date +%Y%m%d%H%M%S)"
        cp -f "${dest}" "${backup}"
        log_verbose "Backed up existing file: ${dest} -> ${backup}"
    fi

    cp -f "${src}" "${dest}"
    log_verbose "Copied: ${src} -> ${dest}"
}

# Copy directory recursively
copy_dir() {
    local src="$1"
    local dest="$2"
    
    if [[ ! -d "${src}" ]]; then
        log_verbose "Source directory not found: ${src}"
        return 1
    fi

    ensure_dir "$(dirname "${dest}")"
    cp -rf "${src}" "${dest}"
    log_verbose "Copied directory: ${src} -> ${dest}"
}

# =============================================================================
# KDE Configuration Paths
# =============================================================================

# KDE Plasma 6 configuration directories
readonly KDE_CONFIG_DIR="${HOME}/.config"
readonly KDE_DATA_DIR="${HOME}/.local/share"
readonly KDE_CACHE_DIR="${HOME}/.cache"

# =============================================================================
# Category-specific file lists
# =============================================================================

# Get list of shortcut configuration files
get_shortcut_files() {
    echo "${KDE_CONFIG_DIR}/kglobalshortcutsrc"
    echo "${KDE_CONFIG_DIR}/khotkeysrc"
}

# Get list of window management configuration files
get_window_management_files() {
    echo "${KDE_CONFIG_DIR}/kwinrc"
    echo "${KDE_CONFIG_DIR}/kwinrulesrc"
    echo "${KDE_DATA_DIR}/kwin"
}

# Get list of theme configuration files
get_theme_files() {
    echo "${KDE_CONFIG_DIR}/kdeglobals"
    echo "${KDE_CONFIG_DIR}/plasmarc"
    echo "${KDE_CONFIG_DIR}/breezerc"
    echo "${KDE_CONFIG_DIR}/auroraerc"
    echo "${KDE_CONFIG_DIR}/kcminputrc"
    echo "${KDE_CONFIG_DIR}/gtkrc"
    echo "${KDE_CONFIG_DIR}/gtk-3.0/settings.ini"
    echo "${KDE_CONFIG_DIR}/gtk-4.0/settings.ini"
    echo "${KDE_CONFIG_DIR}/kwinrc"  # Contains window decoration settings
    echo "${KDE_DATA_DIR}/color-schemes"
    echo "${KDE_DATA_DIR}/wallpapers"
    echo "${KDE_DATA_DIR}/icons"
    echo "${KDE_DATA_DIR}/plasma/look-and-feel"
    echo "${KDE_DATA_DIR}/aurorae/themes"
    echo "${KDE_DATA_DIR}/plasma/desktoptheme"
}

# Get list of language/locale configuration files
get_language_files() {
    echo "${KDE_CONFIG_DIR}/plasma-localerc"
    echo "${KDE_CONFIG_DIR}/kdeglobals"  # Contains [Locale] section
    echo "${KDE_CONFIG_DIR}/fcitx5"
    echo "${KDE_CONFIG_DIR}/ibus"
}

# Get list of widget configuration files
get_widget_files() {
    echo "${KDE_DATA_DIR}/plasma/org.kde.plasma.desktop-appletsrc"
    echo "${KDE_DATA_DIR}/plasma/org.kde.panel-appletsrc"
    echo "${KDE_DATA_DIR}/plasma/plasmoids"
    echo "${KDE_DATA_DIR}/plasma/layout-templates"
    echo "${KDE_DATA_DIR}/plasma/packages"
}

# Get list of panel configuration files
get_panel_files() {
    echo "${KDE_DATA_DIR}/plasma/org.kde.panel"
    echo "${KDE_CONFIG_DIR}/plasmarc"
}

# Get list of system settings files
get_system_settings_files() {
    echo "${KDE_CONFIG_DIR}/kdeglobals"
    echo "${KDE_CONFIG_DIR}/systemsettingsrc"
    echo "${KDE_CONFIG_DIR}/powerdevilrc"
    echo "${KDE_CONFIG_DIR}/kscreenlockerrc"
    echo "${KDE_CONFIG_DIR}/kded6rc"
    echo "${KDE_CONFIG_DIR}/ksplashrc"
    echo "${KDE_CONFIG_DIR}/startkderc"
    echo "${KDE_CONFIG_DIR}/ksmserverrc"
    echo "${KDE_CONFIG_DIR}/krunnerrc"
    echo "${KDE_CONFIG_DIR}/kwalletrc"
    echo "${KDE_CONFIG_DIR}/baloofilerc"
    echo "${KDE_CONFIG_DIR}/dolphinrc"
    echo "${KDE_CONFIG_DIR}/katerc"
    echo "${KDE_CONFIG_DIR}/konsoleshellrc"
}

# Get files for a specific category
get_category_files() {
    local category="$1"
    case "${category}" in
        shortcuts)       get_shortcut_files ;;
        window_management) get_window_management_files ;;
        themes)          get_theme_files ;;
        languages)       get_language_files ;;
        widgets)         get_widget_files ;;
        panels)          get_panel_files ;;
        system_settings) get_system_settings_files ;;
        *)
            log_warning "Unknown category: ${category}"
            return 1
            ;;
    esac
}

# =============================================================================
# Argument Parsing
# =============================================================================

# Parse common arguments for all scripts
parse_common_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dotfiles-dir|-d)
                DOTFILES_DIR="$2"
                shift 2
                ;;
            --categories|-c)
                CATEGORIES="$2"
                shift 2
                ;;
            --verbose|-v)
                VERBOSE=1
                shift
                ;;
            --help|-h)
                return 0
                ;;
            *)
                shift
                ;;
        esac
    done
}
