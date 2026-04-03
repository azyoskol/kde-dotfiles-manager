#!/usr/bin/env bash
# =============================================================================
# KDE Dotfiles Manager - Backup Script
# Backs up KDE Plasma 6+ configuration files to a dotfiles directory
# =============================================================================

set -euo pipefail

# Load common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration
DOTFILES_DIR="${DEFAULT_DOTFILES_DIR}"
CATEGORIES="${DEFAULT_CATEGORIES}"
PROFILE="${1:-default}"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

# =============================================================================
# Usage
# =============================================================================

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS] [PROFILE]

Backup KDE Plasma 6+ configuration files to dotfiles directory.

Options:
  -d, --dotfiles-dir DIR    Directory to store backups (default: ~/kde-dotfiles)
  -c, --categories CATS     Comma-separated list of categories to backup
                            (default: ${DEFAULT_CATEGORIES})
  -v, --verbose             Enable verbose output
  -h, --help                Show this help message

Categories:
  shortcuts          Keyboard shortcuts and global hotkeys
  themes             Color schemes, icons, cursors, wallpapers
  window_management  KWin rules, virtual desktops, tiling
  languages          Locale, input methods, keyboard layouts
  widgets            Desktop widgets and plasmoids
  panels             Panel layout and configuration
  system_settings    General system settings

Examples:
  $(basename "$0")                           # Backup all categories to 'default' profile
  $(basename "$0") work                      # Backup to 'work' profile
  $(basename "$0") -c shortcuts,themes       # Backup only shortcuts and themes
  $(basename "$0") -d /path/to/dir gaming    # Backup to custom directory
EOF
}

# =============================================================================
# Backup Functions
# =============================================================================

# Backup a single file, preserving directory structure
backup_file() {
    local src="$1"
    local dest_dir="$2"
    local category="$3"

    if [[ ! -e "${src}" ]]; then
        log_verbose "Skipping (not found): ${src}"
        return 0
    fi

    # Create relative path structure in destination
    local relative_path="${src#${HOME}/}"
    local dest="${dest_dir}/${relative_path}"

    if [[ -f "${src}" ]]; then
        ensure_dir "$(dirname "${dest}")"
        cp -f "${src}" "${dest}"
        log_verbose "Backed up file: ${src}"
    elif [[ -d "${src}" ]]; then
        cp -rf "${src}" "${dest}"
        log_verbose "Backed up directory: ${src}"
    fi
}

# Backup a specific category
backup_category() {
    local category="$1"
    local dest_dir="$2"
    local category_dir="${dest_dir}/${category}"

    log_info "Backing up category: ${category}"
    ensure_dir "${category_dir}"

    local files
    files=$(get_category_files "${category}")

    local backed_up=0
    local skipped=0

    while IFS= read -r file; do
        [[ -z "${file}" ]] && continue

        if [[ -e "${file}" ]]; then
            backup_file "${file}" "${dest_dir}" "${category}"
            ((backed_up++))
        else
            log_verbose "Skipping (not found): ${file}"
            ((skipped++))
        fi
    done <<< "${files}"

    log_success "Category '${category}': ${backed_up} items backed up, ${skipped} skipped"
}

# Create a manifest file listing all backed up files
create_manifest() {
    local dest_dir="$1"
    local manifest="${dest_dir}/MANIFEST.md"

    cat > "${manifest}" <<EOF
# KDE Dotfiles Backup Manifest

**Profile:** ${PROFILE}
**Date:** ${TIMESTAMP}
**KDE Version:** $(plasmashell --version 2>/dev/null || echo "Unknown")
**Hostname:** $(hostname)

## Backed Up Categories

EOF

    IFS=',' read -ra CATS <<< "${CATEGORIES}"
    for category in "${CATS[@]}"; do
        echo "- ${category}" >> "${manifest}"
    done

    echo "" >> "${manifest}"
    echo "## File List" >> "${manifest}"
    echo "" >> "${manifest}"

    # List all backed up files
    find "${dest_dir}" -type f ! -name "MANIFEST.md" | while read -r file; do
        local relative="${file#${dest_dir}/}"
        echo "- \`${relative}\`" >> "${manifest}"
    done

    log_success "Manifest created: ${manifest}"
}

# =============================================================================
# Main
# =============================================================================

main() {
    # Parse arguments
    local positional_args=()
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
                usage
                exit 0
                ;;
            -*)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
            *)
                positional_args+=("$1")
                shift
                ;;
        esac
    done

    # Set profile from positional argument
    if [[ ${#positional_args[@]} -gt 0 ]]; then
        PROFILE="${positional_args[0]}"
    fi

    # Expand tilde in path
    DOTFILES_DIR="${DOTFILES_DIR/#\~/$HOME}"
    local backup_dir="${DOTFILES_DIR}/${PROFILE}"

    # Header
    echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║     KDE Dotfiles Manager - Backup        ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
    echo ""
    log_info "Profile: ${PROFILE}"
    log_info "Backup directory: ${backup_dir}"
    log_info "Categories: ${CATEGORIES}"
    echo ""

    # Check KDE Plasma
    check_kde_plasma || true

    # Create backup directory
    ensure_dir "${backup_dir}"

    # Backup each category
    local total_backed_up=0
    IFS=',' read -ra CATS <<< "${CATEGORIES}"
    for category in "${CATS[@]}"; do
        category=$(echo "${category}" | xargs) # Trim whitespace
        backup_category "${category}" "${backup_dir}"
        ((total_backed_up++))
    done

    # Create manifest
    create_manifest "${backup_dir}"

    # Summary
    echo ""
    echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║           Backup Complete                ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
    echo ""
    log_success "Backup saved to: ${backup_dir}"
    log_info "Profile: ${PROFILE}"
    log_info "Categories backed up: ${total_backed_up}"
}

main "$@"
