#!/bin/bash

# Forgejo MCP Release Script
# Automates the release process for the Forgejo MCP server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="forgejo-mcp"
GITHUB_REPO="kunde21/forgejo-mcp"
DOCKER_IMAGE="kunde21/forgejo-mcp"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
}

# Check if working directory is clean
check_clean_working_directory() {
    if ! git diff --quiet || ! git diff --staged --quiet; then
        log_error "Working directory is not clean. Please commit or stash changes."
        exit 1
    fi
}

# Get current version from git tags
get_current_version() {
    local version
    version=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    echo "$version"
}

# Determine next version
get_next_version() {
    local current_version="$1"
    local version_type="$2"

    # Remove 'v' prefix if present
    current_version="${current_version#v}"

    # Split version into components
    IFS='.' read -r major minor patch <<< "$current_version"

    case "$version_type" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            log_error "Invalid version type: $version_type"
            echo "Usage: $0 [major|minor|patch]"
            exit 1
            ;;
    esac

    echo "v$major.$minor.$patch"
}

# Create and push git tag
create_git_tag() {
    local version="$1"

    log_info "Creating git tag: $version"
    git tag -a "$version" -m "Release $version"
    git push origin "$version"

    log_success "Git tag created and pushed: $version"
}

# Build binaries for release
build_release_binaries() {
    local version="$1"

    log_info "Building release binaries for version $version"

    # Create bin directory
    mkdir -p bin

    # Build for multiple platforms
    local platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

    for platform in "${platforms[@]}"; do
        IFS='/' read -r os arch <<< "$platform"

        local output_name="bin/${PROJECT_NAME}"
        if [ "$os" = "windows" ]; then
            output_name="${output_name}.exe"
        fi

        log_info "Building for $os/$arch"
        GOOS="$os" GOARCH="$arch" go build \
            -ldflags "-X main.version=$version -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
            -o "$output_name" \
            ./cmd

        # Create archive
        if [ "$os" = "windows" ]; then
            zip -r "bin/${PROJECT_NAME}-${version}-${os}-${arch}.zip" "$output_name"
        else
            tar -czf "bin/${PROJECT_NAME}-${version}-${os}-${arch}.tar.gz" "$output_name"
        fi

        rm "$output_name"
    done

    log_success "Release binaries built successfully"
}

# Generate checksums
generate_checksums() {
    log_info "Generating checksums for release artifacts"

    cd bin
    sha256sum *.tar.gz *.zip > checksums.txt 2>/dev/null || true
    cd ..

    log_success "Checksums generated: bin/checksums.txt"
}

# Create GitHub release
create_github_release() {
    local version="$1"

    if ! command -v gh &> /dev/null; then
        log_warning "GitHub CLI not found. Skipping GitHub release creation."
        log_info "Manual release creation required at: https://github.com/$GITHUB_REPO/releases/new"
        return
    fi

    log_info "Creating GitHub release: $version"

    # Create release notes
    local release_notes
    release_notes=$(git log --pretty=format:"%h %s" "$(git describe --tags --abbrev=0)..HEAD" | head -10)

    gh release create "$version" \
        --title "Release $version" \
        --notes "$release_notes" \
        bin/*

    log_success "GitHub release created: $version"
}

# Build and push Docker image
build_docker_image() {
    local version="$1"

    if ! command -v docker &> /dev/null; then
        log_warning "Docker not found. Skipping Docker image build."
        return
    fi

    log_info "Building Docker image: $DOCKER_IMAGE:$version"

    docker build -t "$DOCKER_IMAGE:$version" .
    docker tag "$DOCKER_IMAGE:$version" "$DOCKER_IMAGE:latest"

    # Push to Docker Hub (if configured)
    if [ -n "$DOCKER_USERNAME" ] && [ -n "$DOCKER_PASSWORD" ]; then
        log_info "Pushing Docker image to registry"
        echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
        docker push "$DOCKER_IMAGE:$version"
        docker push "$DOCKER_IMAGE:latest"
        log_success "Docker image pushed successfully"
    else
        log_warning "Docker credentials not configured. Image built but not pushed."
    fi
}

# Main release process
main() {
    local version_type="${1:-patch}"
    local dry_run="${2:-false}"

    log_info "Starting Forgejo MCP release process"
    log_info "Version type: $version_type"

    # Pre-flight checks
    check_git_repo
    check_clean_working_directory

    # Get version information
    local current_version
    current_version=$(get_current_version)
    local next_version
    next_version=$(get_next_version "$current_version" "$version_type")

    log_info "Current version: $current_version"
    log_info "Next version: $next_version"

    if [ "$dry_run" = "true" ]; then
        log_info "DRY RUN - Would create release: $next_version"
        exit 0
    fi

    # Execute release steps
    create_git_tag "$next_version"
    build_release_binaries "$next_version"
    generate_checksums
    create_github_release "$next_version"
    build_docker_image "$next_version"

    log_success "Release $next_version completed successfully!"
    log_info "Release artifacts available in: bin/"
    log_info "GitHub release: https://github.com/$GITHUB_REPO/releases/tag/$next_version"
}

# Parse command line arguments
DRY_RUN=false
VERSION_TYPE="patch"

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        major|minor|patch)
            VERSION_TYPE="$1"
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [major|minor|patch] [--dry-run]"
            echo ""
            echo "Arguments:"
            echo "  major|minor|patch    Version increment type (default: patch)"
            echo "  --dry-run           Show what would be done without making changes"
            echo ""
            echo "Examples:"
            echo "  $0 patch           # Create patch release"
            echo "  $0 minor --dry-run # Preview minor release"
            exit 0
            ;;
        *)
            log_error "Unknown argument: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Run main function
main "$VERSION_TYPE" "$DRY_RUN"