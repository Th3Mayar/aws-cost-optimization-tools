#!/bin/bash

set -e

echo "=========================================="
echo "AWS Cost Optimization Tools - Setup"
echo "=========================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed."
    echo ""
    echo "Please install Go first:"
    echo ""
    echo "  Ubuntu/Debian:"
    echo "    sudo snap install go --classic"
    echo ""
    echo "  Or download from: https://go.dev/dl/"
    echo ""
    exit 1
fi

GO_VERSION=$(go version)
echo "✅ Go detected: $GO_VERSION"
echo ""

# Download dependencies
echo "📚 Downloading dependencies..."
go mod download
go mod tidy
echo "✅ Dependencies installed"
echo ""

# Build binary
echo "📦 Building binary..."
mkdir -p bin
go build -o bin/cost-optimization ./cmd/cost-optimization
echo "✅ Binary built: bin/cost-optimization"
echo ""

# Test the binary
echo "🧪 Testing binary..."
./bin/cost-optimization --version
echo ""

echo "=========================================="
echo "✅ Setup completed successfully!"
echo "=========================================="
echo ""
echo "You can use the binary in the following ways:"
echo ""
echo "  1. Directly:"
echo "     ./bin/cost-optimization start"
echo ""
echo "  2. Install globally:"
echo "     sudo cp bin/cost-optimization /usr/local/bin/"
echo "     cost-optimization start"
echo ""
echo "  3. Or use make:"
echo "     make install"
echo ""
echo "For more information:"
echo "  ./bin/cost-optimization --help"
echo "  cat README.md"
echo ""
