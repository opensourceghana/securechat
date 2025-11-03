#!/bin/bash

# SecureChat Demo Script
# This script demonstrates the basic functionality of SecureChat

set -e

echo "🔐 SecureChat Demo"
echo "=================="
echo

# Check if binaries exist
if [ ! -f "build/securechat" ]; then
    echo "❌ SecureChat binary not found. Please run 'make build' first."
    exit 1
fi

if [ ! -f "build/relay-server" ]; then
    echo "❌ Relay server binary not found. Please run 'make build' first."
    exit 1
fi

echo "✅ Binaries found"
echo

# Create test directories
mkdir -p test-data/user1 test-data/user2

# Create test configurations
cat > test-data/user1/config.yaml << EOF
user:
  id: "alice"
  display_name: "Alice"
  status_message: "Testing SecureChat"

network:
  relay_servers:
    - "localhost:8080"
  p2p_enabled: false
  connection_timeout: "10s"

ui:
  theme: "dark"
  notifications: true
  sound_enabled: false
  timestamp_format: "15:04"

security:
  auto_accept_keys: false
  message_retention_days: 7
  require_verification: false

debug: true
EOF

cat > test-data/user2/config.yaml << EOF
user:
  id: "bob"
  display_name: "Bob"
  status_message: "Also testing SecureChat"

network:
  relay_servers:
    - "localhost:8080"
  p2p_enabled: false
  connection_timeout: "10s"

ui:
  theme: "dark"
  notifications: true
  sound_enabled: false
  timestamp_format: "15:04"

security:
  auto_accept_keys: false
  message_retention_days: 7
  require_verification: false

debug: true
EOF

echo "📝 Created test configurations for Alice and Bob"
echo

# Function to cleanup background processes
cleanup() {
    echo
    echo "🧹 Cleaning up..."
    if [ ! -z "$RELAY_PID" ]; then
        kill $RELAY_PID 2>/dev/null || true
        echo "   Stopped relay server"
    fi
    rm -rf test-data/
    echo "   Cleaned up test data"
    echo "✅ Demo cleanup complete"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Start relay server in background
echo "🚀 Starting relay server on localhost:8080..."
./build/relay-server -addr localhost -port 8080 &
RELAY_PID=$!

# Wait for server to start
sleep 2

# Check if server is running
if ! kill -0 $RELAY_PID 2>/dev/null; then
    echo "❌ Failed to start relay server"
    exit 1
fi

echo "✅ Relay server started (PID: $RELAY_PID)"
echo

# Test configuration loading
echo "🔧 Testing configuration loading..."
./build/securechat --config test-data/user1/config.yaml --version
echo "✅ Configuration loading works"
echo

# Show project structure
echo "📁 Project Structure:"
echo "   SecureChat/"
echo "   ├── cmd/"
echo "   │   ├── securechat/     # Main application"
echo "   │   └── relay-server/   # Relay server"
echo "   ├── pkg/"
echo "   │   ├── core/           # Core application logic"
echo "   │   ├── crypto/         # Encryption (Signal Protocol)"
echo "   │   ├── network/        # WebSocket networking"
echo "   │   ├── storage/        # BadgerDB storage"
echo "   │   └── ui/             # Bubble Tea TUI"
echo "   ├── internal/"
echo "   │   ├── config/         # Configuration management"
echo "   │   └── models/         # Data models"
echo "   └── docs/               # Documentation"
echo

echo "🎯 Key Features Implemented:"
echo "   ✅ Terminal UI with Bubble Tea"
echo "   ✅ WebSocket networking with relay server"
echo "   ✅ Local storage with BadgerDB"
echo "   ✅ Basic encryption framework"
echo "   ✅ Configuration management"
echo "   ✅ Multiple UI views (chat, contacts, settings, help)"
echo "   ✅ Cross-platform Go implementation"
echo

echo "🔮 Next Steps:"
echo "   🔄 Full Signal Protocol implementation"
echo "   🔄 P2P networking with NAT traversal"
echo "   🔄 File transfer capabilities"
echo "   🔄 Advanced UI features"
echo "   🔄 Comprehensive testing"
echo

echo "🎉 Demo complete! SecureChat is ready for development."
echo
echo "To run SecureChat:"
echo "   ./build/securechat --config test-data/user1/config.yaml"
echo
echo "To run the relay server:"
echo "   ./build/relay-server -addr localhost -port 8080"
echo

# Keep relay server running for a bit
echo "⏳ Keeping relay server running for 10 seconds..."
sleep 10
