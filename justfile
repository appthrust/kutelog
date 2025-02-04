# デフォルトのバージョン
version := "dev"

# バージョンを指定してビルド
build: clean
    go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog ./cmd/main.go

# クロスビルド
build-all: clean
    # Linux (amd64, arm64)
    GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-linux-amd64 ./cmd/main.go
    GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-linux-arm64 ./cmd/main.go
    # macOS (amd64, arm64)
    GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-darwin-amd64 ./cmd/main.go
    GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-darwin-arm64 ./cmd/main.go
    # Windows (amd64)
    GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/appthrust/kutelog/pkg/version.Version={{version}}" -o dist/kutelog-windows-amd64.exe ./cmd/main.go

# distディレクトリをクリーンアップ
clean:
    rm -rf dist
    mkdir -p dist
