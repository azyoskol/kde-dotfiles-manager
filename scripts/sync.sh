#!/usr/bin/env bash
# =============================================================================
# KDE Dotfiles Manager - Sync Script
# Synchronizes dotfiles with a Git repository
# =============================================================================

set -euo pipefail

# Load common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration
DOTFILES_DIR="${DEFAULT_DOTFILES_DIR}"
GIT_REMOTE="origin"
GIT_BRANCH="main"
GIT_REPO=""
COMMIT_MESSAGE=""

# =============================================================================
# Usage
# =============================================================================

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Synchronize KDE dotfiles with a Git repository.

Options:
  -d, --dotfiles-dir DIR    Directory containing dotfiles (default: ~/kde-dotfiles)
  -r, --remote URL          Git remote repository URL
  -b, --branch BRANCH       Git branch to use (default: main)
  -m, --message MSG         Commit message (default: auto-generated)
  --init                    Initialize a new git repository
  --add                     Stage all changes
  --commit                  Commit staged changes
  --push                    Push to remote repository
  --pull                    Pull from remote repository
  --status                  Show git status
  --clone                   Clone remote repository
  -v, --verbose             Enable verbose output
  -h, --help                Show this help message

Examples:
  $(basename "$0") --init                              # Initialize git repo
  $(basename "$0") -r git@github.com:user/dotfiles.git # Set remote
  $(basename "$0") --push                              # Push to remote
  $(basename "$0") --pull                              # Pull from remote
  $(basename "$0") --clone -r URL                      # Clone repository
  $(basename "$0") --status                            # Show status
EOF
}

# =============================================================================
# Git Functions
# =============================================================================

# Check if directory is a git repository
is_git_repo() {
    [[ -d "${DOTFILES_DIR}/.git" ]]
}

# Initialize git repository
git_init() {
    log_info "Initializing git repository in ${DOTFILES_DIR}"
    ensure_dir "${DOTFILES_DIR}"
    
    git -C "${DOTFILES_DIR}" init
    log_success "Git repository initialized"
    
    # Create .gitignore
    cat > "${DOTFILES_DIR}/.gitignore" <<'EOF'
# Ignore backup files
*.bak.*

# Ignore sensitive files
*.key
*.pem
*.crt

# Ignore cache
.cache/
*.cache

# Ignore temporary files
*.tmp
*.swp
*~
EOF
    log_info "Created .gitignore"
}

# Set remote repository
git_set_remote() {
    local url="$1"
    
    if is_git_repo; then
        # Remove existing remote if present
        git -C "${DOTFILES_DIR}" remote remove "${GIT_REMOTE}" 2>/dev/null || true
        git -C "${DOTFILES_DIR}" remote add "${GIT_REMOTE}" "${url}"
        log_success "Remote '${GIT_REMOTE}' set to: ${url}"
    else
        log_error "Not a git repository. Run --init first."
        return 1
    fi
}

# Stage all changes
git_add() {
    if ! is_git_repo; then
        log_error "Not a git repository. Run --init first."
        return 1
    fi
    
    log_info "Staging all changes..."
    git -C "${DOTFILES_DIR}" add -A
    log_success "Changes staged"
}

# Commit changes
git_commit() {
    if ! is_git_repo; then
        log_error "Not a git repository. Run --init first."
        return 1
    fi
    
    # Check if there are changes to commit
    if [[ -z "$(git -C "${DOTFILES_DIR}" status --porcelain)" ]]; then
        log_info "No changes to commit"
        return 0
    fi
    
    local msg="${COMMIT_MESSAGE:-Update KDE dotfiles - $(date '+%Y-%m-%d %H:%M:%S')}"
    
    git -C "${DOTFILES_DIR}" commit -m "${msg}"
    log_success "Committed: ${msg}"
}

# Push to remote
git_push() {
    if ! is_git_repo; then
        log_error "Not a git repository. Run --init first."
        return 1
    fi
    
    # Check if remote is configured
    if ! git -C "${DOTFILES_DIR}" remote get-url "${GIT_REMOTE}" &>/dev/null; then
        log_error "Remote '${GIT_REMOTE}' is not configured. Use -r to set it."
        return 1
    fi
    
    log_info "Pushing to ${GIT_REMOTE}/${GIT_BRANCH}..."
    git -C "${DOTFILES_DIR}" push -u "${GIT_REMOTE}" "${GIT_BRANCH}"
    log_success "Pushed to ${GIT_REMOTE}/${GIT_BRANCH}"
}

# Pull from remote
git_pull() {
    if ! is_git_repo; then
        log_error "Not a git repository. Run --init first."
        return 1
    fi
    
    # Check if remote is configured
    if ! git -C "${DOTFILES_DIR}" remote get-url "${GIT_REMOTE}" &>/dev/null; then
        log_error "Remote '${GIT_REMOTE}' is not configured. Use -r to set it."
        return 1
    fi
    
    log_info "Pulling from ${GIT_REMOTE}/${GIT_BRANCH}..."
    git -C "${DOTFILES_DIR}" pull "${GIT_REMOTE}" "${GIT_BRANCH}"
    log_success "Pulled from ${GIT_REMOTE}/${GIT_BRANCH}"
}

# Show git status
git_status() {
    if ! is_git_repo; then
        log_warning "Not a git repository"
        return 1
    fi
    
    echo ""
    echo -e "${BOLD}Git Status:${NC}"
    echo ""
    git -C "${DOTFILES_DIR}" status
    echo ""
    
    echo -e "${BOLD}Recent Commits:${NC}"
    echo ""
    git -C "${DOTFILES_DIR}" log --oneline -10
    echo ""
}

# Clone repository
git_clone() {
    if [[ -z "${GIT_REPO}" ]]; then
        log_error "No repository URL provided. Use -r to specify."
        return 1
    fi
    
    # Remove existing directory if it exists
    if [[ -d "${DOTFILES_DIR}" ]]; then
        log_warning "Removing existing directory: ${DOTFILES_DIR}"
        rm -rf "${DOTFILES_DIR}"
    fi
    
    # Create parent directory
    ensure_dir "$(dirname "${DOTFILES_DIR}")"
    
    log_info "Cloning ${GIT_REPO} to ${DOTFILES_DIR}..."
    git clone "${GIT_REPO}" "${DOTFILES_DIR}"
    log_success "Repository cloned"
}

# Full sync: add, commit, push
git_sync() {
    git_add
    git_commit
    git_push
}

# =============================================================================
# Main
# =============================================================================

main() {
    local action=""
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dotfiles-dir|-d)
                DOTFILES_DIR="$2"
                shift 2
                ;;
            --remote|-r)
                GIT_REPO="$2"
                shift 2
                ;;
            --branch|-b)
                GIT_BRANCH="$2"
                shift 2
                ;;
            --message|-m)
                COMMIT_MESSAGE="$2"
                shift 2
                ;;
            --init)
                action="init"
                shift
                ;;
            --add)
                action="add"
                shift
                ;;
            --commit)
                action="commit"
                shift
                ;;
            --push)
                action="push"
                shift
                ;;
            --pull)
                action="pull"
                shift
                ;;
            --status)
                action="status"
                shift
                ;;
            --clone)
                action="clone"
                shift
                ;;
            --sync)
                action="sync"
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

    # Header
    echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║     KDE Dotfiles Manager - Sync          ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
    echo ""
    log_info "Dotfiles directory: ${DOTFILES_DIR}"
    
    if [[ -n "${GIT_REPO}" ]]; then
        log_info "Remote repository: ${GIT_REPO}"
    fi
    echo ""

    # Set remote if provided
    if [[ -n "${GIT_REPO}" ]] && is_git_repo; then
        git_set_remote "${GIT_REPO}"
    fi

    # Execute action
    case "${action}" in
        init)
            git_init
            ;;
        add)
            git_add
            ;;
        commit)
            git_commit
            ;;
        push)
            git_push
            ;;
        pull)
            git_pull
            ;;
        status)
            git_status
            ;;
        clone)
            git_clone
            ;;
        sync)
            git_sync
            ;;
        "")
            # Default: full sync
            if ! is_git_repo; then
                log_info "No git repository found. Initializing..."
                git_init
            fi
            git_sync
            ;;
    esac
}

main "$@"
