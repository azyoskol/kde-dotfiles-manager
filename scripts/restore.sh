#!/usr/bin/env bash
# =============================================================================
# KDE Dotfiles Manager - Restore Script
# Restores KDE Plasma 6+ configuration files from a dotfiles backup
# =============================================================================

set -euo pipefail

# Load common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration
DOTFILES_DIR="${DEFAULT_DOTFILES_DIR}"
CATEGORIES="${DEFAULT_CATEGORIES}"
PROFILE="default"
INTERACTIVE=0
FORCE=0

# =============================================================================
# Usage
# =============================================================================

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Restore KDE Plasma 6+ configuration files from a dotfiles backup.

Options:
  -d, --dotfiles-dir DIR    Directory containing backups (default: ~/kde-dotfiles)
  -p, --profile PROFILE     Backup profile to restore from (default: default)
  -c, --categories CATS     Comma-separated list of categories to restore
                            (default: ${DEFAULT_CATEGORIES})
  -i, --interactive         Prompt before restoring each category
  -f, --force               Skip confirmation prompt
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
  $(basename "$0")                           # Restore all categories from 'default'
  $(basename "$0") -p work                   # Restore from 'work' profile
  $(basename "$0") -c shortcuts,themes       # Restore only shortcuts and themes
  $(basename "$0") -i                        # Interactive restore with prompts
EOF
}

# =============================================================================
# Restore Functions
# =============================================================================

# Restore a single file from backup
restore_file() {
    local backup_file="$1"
    local original_path="$2"

    if [[ ! -e "${backup_file}" ]]; then
        log_verbose "Backup not found: ${backup_file}"
        return 0
    fi

    # Create parent directory if needed
    ensure_dir "$(dirname "${original_path}")"

    if [[ -f "${backup_file}" ]]; then
        copy_with_backup "${backup_file}" "${original_path}"
        log_verbose "Restored file: ${original_path}"
    elif [[ -d "${backup_file}" ]]; then
        copy_dir "${backup_file}" "${original_path}"
        log_verbose "Restored directory: ${original_path}"
    fi
}

# Restore a specific category
restore_category() {
    local category="$1"
    local backup_dir="$2"
    local category_backup="${backup_dir}/${category}"

    # Check if category backup exists
    if [[ ! -d "${category_backup}" ]] && [[ ! -e "${backup_dir}" ]]; then
        log_warning "No backup found for category: ${category}"
        return 1
    fi

    log_info "Restoring category: ${category}"

    local files
    files=$(get_category_files "${category}")

    local restored=0
    local skipped=0

    while IFS= read -r file; do
        [[ -z "${file}" ]] && continue

        # Find corresponding backup file
        local relative_path="${file#${HOME}/}"
        local backup_path="${backup_dir}/${relative_path}"

        if [[ -e "${backup_path}" ]]; then
            restore_file "${backup_path}" "${file}"
            ((restored++))
        else
            log_verbose "No backup for: ${file}"
            ((skipped++))
        fi
    done <<< "${files}"

    log_success "Category '${category}': ${restored} items restored, ${skipped} skipped"
}

# Prompt user for confirmation
prompt_confirm() {
    local message="$1"
    if [[ "${INTERACTIVE}" -eq 1 ]]; then
        echo -ne "${YELLOW}${message} [y/N]${NC} "
        read -r response
        if [[ "${response}" =~ ^[Yy]$ ]]; then
            return 0
        fi
        return 1
    fi
    return 0
}

# List available profiles
list_profiles() {
    local profiles_dir="${DOTFILES_DIR}"
    
    if [[ ! -d "${profiles_dir}" ]]; then
        log_warning "No dotfiles directory found at: ${profiles_dir}"
        return 1
    fi

    echo -e "${BOLD}Available backup profiles:${NC}"
    echo ""
    
    for profile in "${profiles_dir}"/*/; do
        if [[ -d "${profile}" ]]; then
            local name
            name=$(basename "${profile}")
            local manifest="${profile}/MANIFEST.md"
            local date="unknown"
            
            if [[ -f "${manifest}" ]]; then
                date=$(grep -m1 "**Date:**" "${manifest}" 2>/dev/null | cut -d'**' -f4 || echo "unknown")
            fi
            
            echo "  - ${name} (${date})"
        fi
    done
    echo ""
}

# =============================================================================
# Main
# =============================================================================

main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dotfiles-dir|-d)
                DOTFILES_DIR="$2"
                shift 2
                ;;
            --profile|-p)
                PROFILE="$2"
                shift 2
                ;;
            --categories|-c)
                CATEGORIES="$2"
                shift 2
                ;;
            --interactive|-i)
                INTERACTIVE=1
                shift
                ;;
            --force|-f)
                FORCE=1
                shift
                ;;
            --verbose|-v)
                VERBOSE=1
                shift
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            --list-profiles|-l)
                list_profiles
                exit 0
                ;;
            -*)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
            *)
                shift
                ;;
        esac
    done

    # Expand tilde in path
    DOTFILES_DIR="${DOTFILES_DIR/#\~/$HOME}"
    local backup_dir="${DOTFILES_DIR}/${PROFILE}"

    # Header
    echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║     KDE Dotfiles Manager - Restore       ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
    echo ""
    log_info "Profile: ${PROFILE}"
    log_info "Backup directory: ${backup_dir}"
    log_info "Categories: ${CATEGORIES}"
    echo ""

    # Check if backup exists
    if [[ ! -d "${backup_dir}" ]]; then
        log_error "Backup profile not found: ${backup_dir}"
        echo ""
        list_profiles || true
        exit 1
    fi

    # Confirmation prompt
    if [[ "${FORCE}" -ne 1 ]]; then
        echo -e "${YELLOW}WARNING: This will overwrite your current KDE configuration!${NC}"
        echo -e "${YELLOW}Make sure you have a backup of your current settings.${NC}"
        echo ""
        
        if ! prompt_confirm "Continue with restore?"; then
            log_info "Restore cancelled by user"
            exit 0
        fi
    fi

    # Stop KDE services before restoring
    stop_kde_services

    # Restore each category
    local total_restored=0
    IFS=',' read -ra CATS <<< "${CATEGORIES}"
    for category in "${CATS[@]}"; do
        category=$(echo "${category}" | xargs) # Trim whitespace
        
        if prompt_confirm "Restore category: ${category}?"; then
            restore_category "${category}" "${backup_dir}" || true
            ((total_restored++))
        fi
    done

    # Restart KDE services
    start_kde_services

    # Summary
    echo ""
    echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║           Restore Complete               ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
    echo ""
    log_success "Restored ${total_restored} categories from profile: ${PROFILE}"
    log_info "You may need to log out and log back in for all changes to take effect"
    
    # Check for custom widgets
    local widgets_backup="${backup_dir}/widgets/plasma/plasmoids"
    if [[ -d "${widgets_backup}" ]]; then
        echo ""
        log_info "Custom widgets found in backup. They will be installed automatically."
        log_info "If you encounter any issues, you can manually install widgets from:"
        log_info "  ${widgets_backup}"
    fi
}

main "$@"
