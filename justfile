# Default version
version := "dev"

# Build with specified version
build: clean
    go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog ./cmd/main.go

# Cross-platform build
build-all: clean
    # Linux (amd64, arm64)
    GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-linux-amd64 ./cmd/main.go
    GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-linux-arm64 ./cmd/main.go
    # macOS (amd64, arm64)
    GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-darwin-amd64 ./cmd/main.go
    GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-darwin-arm64 ./cmd/main.go
    # Windows (amd64)
    GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-windows-amd64.exe ./cmd/main.go

# Install to /usr/local/bin
install: build
    sudo cp dist/kutelog /usr/local/bin/kutelog

# Clean up dist directory
clean:
    rm -rf dist
    mkdir -p dist
