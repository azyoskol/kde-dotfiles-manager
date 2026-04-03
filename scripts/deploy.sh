#!/usr/bin/env bash
# =============================================================================
# KDE Dotfiles Manager - Deploy Script
# Deploys saved KDE configurations to the current system
# =============================================================================

set -euo pipefail

# Load common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration
DOTFILES_DIR="${DEFAULT_DOTFILES_DIR}"
CATEGORIES="${DEFAULT_CATEGORIES}"
PROFILE="default"
DRY_RUN=0
FORCE=0

# =============================================================================
# Usage
# =============================================================================

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Deploy KDE Plasma 6+ configurations from dotfiles to the current system.

Options:
  -d, --dotfiles-dir DIR    Directory containing dotfiles (default: ~/kde-dotfiles)
  -p, --profile PROFILE     Profile to deploy (default: default)
  -c, --categories CATS     Comma-separated list of categories to deploy
                            (default: ${DEFAULT_CATEGORIES})
  --dry-run                 Show what would be deployed without making changes
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
  $(basename "$0")                           # Deploy all categories from 'default'
  $(basename "$0") -p work                   # Deploy 'work' profile
  $(basename "$0") -c shortcuts,themes       # Deploy only shortcuts and themes
  $(basename "$0") --dry-run                 # Preview what would be deployed
EOF
}

# =============================================================================
# Deploy Functions
# =============================================================================

# Deploy a single file (dry-run aware)
deploy_file() {
    local backup_file="$1"
    local dest="$2"

    if [[ ! -e "${backup_file}" ]]; then
        log_verbose "Backup not found: ${backup_file}"
        return 0
    fi

    if [[ "${DRY_RUN}" -eq 1 ]]; then
        log_info "[DRY RUN] Would deploy: ${backup_file} -> ${dest}"
        return 0
    fi

    copy_with_backup "${backup_file}" "${dest}"
    log_verbose "Deployed: ${backup_file} -> ${dest}"
}

# Deploy a specific category
deploy_category() {
    local category="$1"
    local backup_dir="$2"

    log_info "Deploying category: ${category}"

    local files
    files=$(get_category_files "${category}")

    local deployed=0
    local skipped=0

    while IFS= read -r file; do
        [[ -z "${file}" ]] && continue

        # Find corresponding backup file
        local relative_path="${file#${HOME}/}"
        local backup_path="${backup_dir}/${relative_path}"

        if [[ -e "${backup_path}" ]]; then
            deploy_file "${backup_path}" "${file}"
            ((deployed++))
        else
            log_verbose "No backup for: ${file}"
            ((skipped++))
        fi
    done <<< "${files}"

    if [[ "${DRY_RUN}" -eq 1 ]]; then
        log_info "[DRY RUN] Category '${category}': ${deployed} files would be deployed"
    else
        log_success "Category '${category}': ${deployed} files deployed, ${skipped} skipped"
    fi
}

# Show deployment summary
show_summary() {
    local backup_dir="$1"

    echo ""
    echo -e "${BOLD}Deployment Summary:${NC}"
    echo ""

    IFS=',' read -ra CATS <<< "${CATEGORIES}"
    for category in "${CATS[@]}"; do
        category=$(echo "${category}" | xargs)
        local count=0

        local files
        files=$(get_category_files "${category}")
        while IFS= read -r file; do
            [[ -z "${file}" ]] && continue
            local relative_path="${file#${HOME}/}"
            local backup_path="${backup_dir}/${relative_path}"
            if [[ -e "${backup_path}" ]]; then
                ((count++))
            fi
        done <<< "${files}"

        local status="${GREEN}${count} files${NC}"
        if [[ ${count} -eq 0 ]]; then
            status="${RED}no files found${NC}"
        fi

        echo "  ${category}: ${status}"
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
            --dry-run)
                DRY_RUN=1
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
    if [[ "${DRY_RUN}" -eq 1 ]]; then
        echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
        echo -e "${BOLD}║  KDE Dotfiles Manager - Deploy (DRY RUN) ║${NC}"
        echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
    else
        echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
        echo -e "${BOLD}║     KDE Dotfiles Manager - Deploy        ║${NC}"
        echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
    fi
    echo ""
    log_info "Profile: ${PROFILE}"
    log_info "Source directory: ${backup_dir}"
    log_info "Categories: ${CATEGORIES}"
    if [[ "${DRY_RUN}" -eq 1 ]]; then
        log_warning "DRY RUN mode - no changes will be made"
    fi
    echo ""

    # Check if profile exists
    if [[ ! -d "${backup_dir}" ]]; then
        log_error "Profile not found: ${backup_dir}"
        exit 1
    fi

    # Show summary
    show_summary "${backup_dir}"

    # Confirmation prompt
    if [[ "${DRY_RUN}" -ne 1 ]] && [[ "${FORCE}" -ne 1 ]]; then
        echo -e "${YELLOW}WARNING: This will overwrite your current KDE configuration!${NC}"
        echo -e "${YELLOW}Your current files will be backed up with .bak.TIMESTAMP extension.${NC}"
        echo ""
        
        echo -ne "${YELLOW}Continue with deployment? [y/N]${NC} "
        read -r response
        if [[ ! "${response}" =~ ^[Yy]$ ]]; then
            log_info "Deployment cancelled by user"
            exit 0
        fi
    fi

    # Stop KDE services before deploying (not in dry-run)
    if [[ "${DRY_RUN}" -ne 1 ]]; then
        stop_kde_services
    fi

    # Deploy each category
    local total_deployed=0
    IFS=',' read -ra CATS <<< "${CATEGORIES}"
    for category in "${CATS[@]}"; do
        category=$(echo "${category}" | xargs)
        deploy_category "${category}" "${backup_dir}"
        ((total_deployed++))
    done

    # Restart KDE services (not in dry-run)
    if [[ "${DRY_RUN}" -ne 1 ]]; then
        start_kde_services
    fi

    # Summary
    echo ""
    if [[ "${DRY_RUN}" -eq 1 ]]; then
        echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
        echo -e "${BOLD}║        Dry Run Complete                  ║${NC}"
        echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
        echo ""
        log_info "No changes were made. Remove --dry-run to apply changes."
    else
        echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
        echo -e "${BOLD}║          Deploy Complete                 ║${NC}"
        echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
        echo ""
        log_success "Deployed ${total_deployed} categories from profile: ${PROFILE}"
        log_info "You may need to log out and log back in for all changes to take effect"
    fi
}

main "$@"
