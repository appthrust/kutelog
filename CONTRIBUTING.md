# 🌸 Contributing to Kutelog

## Development Environment Setup

### Prerequisites
- [Devbox](https://www.jetpack.io/devbox/) - Development environment manager
- [direnv](https://direnv.net/) (recommended for automatic environment activation)

### Setting Up Development Environment

1. Clone the repository:
   ```bash
   git clone https://github.com/appthrust/kutelog.git
   cd kutelog
   ```

2. Enter development shell:
   ```bash
   devbox shell
   ```
   Or if you have direnv installed and enabled:
   ```bash
   direnv allow
   ```
   
   This will make the following tools available:
   - Go
   - Bun
   - Biome
   - just
   - kind
   - kubectl

3. Install dependencies:
   ```bash
   cd pkg/emitters/websocket/static
   bun install
   ```

### Running Development Server

1. Start frontend development server (in a new terminal):
   ```bash
   cd pkg/emitters/websocket/static
   bun run dev  # runs build in watch mode
   ```

2. Start backend server (in another terminal):
   ```bash
   go run cmd/main.go
   ```

3. Access development server in browser:
   - http://localhost:9106

### Building

1. Build frontend:
   ```bash
   cd pkg/emitters/websocket/static
   bun run build
   ```

2. Build backend:
   ```bash
   go build -o kutelog cmd/main.go
   ```

## Release Process

### Creating a New Release

1. Determine the new version number based on changes:
   - Major version: Breaking changes
   - Minor version: New features
   - Patch version: Bug fixes

2. Create and push a new tag:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"  # Replace with actual version
   git push origin v1.2.3
   ```

3. The release process is automated through GitHub Actions:
   - Publish workflow:
     1. Builds binaries for all supported platforms
     2. Creates a GitHub release with the binaries
     3. Generates release notes from commits
   - Homebrew workflow:
     1. Triggered after successful publish
     2. Updates the Homebrew formula with new version and checksums
     3. Creates a pull request in homebrew-tap repository

### Supported Platforms
- macOS (ARM64, AMD64)
- Linux (ARM64, AMD64)
- Windows (AMD64)
