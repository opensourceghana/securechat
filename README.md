# SecureChat

A minimal, secure, terminal-based chat application designed for developers who want distraction-free communication while staying in their terminal workflow.

## Features

- 🔒 **End-to-end encryption** using the Signal Protocol
- 💻 **Terminal-native interface** with Bubble Tea TUI
- 🚀 **Minimal resource usage** - under 50MB RAM
- 🌐 **Peer-to-peer connections** with relay fallback
- 🔔 **Subtle toast notifications** for incoming messages
- 📱 **Cross-platform support** (Linux, macOS, Windows)
- 🎨 **Customizable themes** (dark/light)
- ⌨️ **Keyboard-driven** interface with vim-like bindings

## Quick Start

### Installation

```bash
# Install from source
go install github.com/opensourceghana/securechat/cmd/securechat@latest

# Or download binary from releases
curl -L https://github.com/opensourceghana/securechat/releases/latest/download/securechat-linux-amd64 -o securechat
chmod +x securechat
```

### First Run

```bash
# Start SecureChat
securechat

# Or run with custom config
securechat --config ~/.config/securechat/config.yaml
```

### Basic Usage

1. **Start a chat:** Press `Ctrl+N` to start a new conversation
2. **Add contacts:** Use `Ctrl+A` to add a contact by their user ID
3. **Send messages:** Type your message and press `Enter`
4. **Switch chats:** Use `Ctrl+T` to cycle through open chats
5. **Settings:** Press `Ctrl+,` to open settings

## Security

SecureChat implements the Signal Protocol for end-to-end encryption:

- **Perfect Forward Secrecy:** Past messages remain secure even if keys are compromised
- **Post-Compromise Security:** Future messages are secure after key compromise recovery
- **Metadata Protection:** Minimal data stored on relay servers
- **Identity Verification:** Manual safety number verification

### Safety Numbers

Each contact has a unique safety number that should be verified out-of-band:

```
12345 67890 12345 67890 12345 67890
```

Compare this number with your contact via voice call or in person to ensure secure communication.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client A      │    │  Relay Server   │    │   Client B      │
│                 │    │   (Optional)    │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │     TUI     │ │    │ │   Router    │ │    │ │     TUI     │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │   Crypto    │ │◄──►│ │  Message    │ │◄──►│ │   Crypto    │ │
│ └─────────────┘ │    │ │  Relay      │ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ └─────────────┘ │    │ ┌─────────────┐ │
│ │  Network    │ │    │ ┌─────────────┐ │    │ │  Network    │ │
│ └─────────────┘ │    │ │   Storage   │ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ └─────────────┘ │    │ ┌─────────────┐ │
│ │   Storage   │ │    └─────────────────┘    │ │   Storage   │ │
│ └─────────────┘ │                           │ └─────────────┘ │
└─────────────────┘                           └─────────────────┘
```

## Configuration

SecureChat looks for configuration in the following locations:

1. `~/.config/securechat/config.yaml`
2. `~/.securechat.yaml`
3. `./config.yaml`

### Example Configuration

```yaml
user:
  display_name: "Alice Cooper"
  status_message: "Working on SecureChat"

network:
  relay_servers:
    - "relay1.securechat.dev:8080"
    - "relay2.securechat.dev:8080"
  p2p_enabled: true
  connection_timeout: "30s"

ui:
  theme: "dark"
  notifications: true
  sound_enabled: false
  timestamp_format: "15:04"

security:
  auto_accept_keys: false
  message_retention_days: 30
  export_keys_path: "~/.config/securechat/keys"
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/opensourceghana/securechat.git
cd secure_chat

# Install dependencies
go mod download

# Build
go build -o securechat cmd/securechat/main.go

# Run tests
go test ./...

# Run with race detection
go run -race cmd/securechat/main.go
```

### Project Structure

```
secure_chat/
├── cmd/
│   └── securechat/          # Main application entry point
├── pkg/
│   ├── crypto/              # Encryption and key management
│   ├── network/             # Network protocols and connections
│   ├── storage/             # Local data storage
│   └── ui/                  # Terminal user interface
├── internal/
│   ├── config/              # Configuration management
│   └── models/              # Data models and types
├── docs/                    # Documentation
├── examples/                # Example configurations and scripts
└── scripts/                 # Build and deployment scripts
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Security Reporting

Please report security vulnerabilities to security@securechat.dev. We will respond within 24 hours and coordinate disclosure.

## Keyboard Shortcuts

### Global
- `Ctrl+N` - New chat
- `Ctrl+T` - Switch chat tabs
- `Ctrl+W` - Close current chat
- `Ctrl+Q` - Quit application
- `Ctrl+,` - Open settings
- `Ctrl+/` - Show help

### Chat
- `Enter` - Send message
- `Shift+Enter` - New line in message
- `Ctrl+L` - Clear chat history
- `Ctrl+F` - Search messages
- `Up/Down` - Navigate message history

### Contacts
- `Ctrl+A` - Add contact
- `Ctrl+E` - Edit contact
- `Del` - Remove contact
- `Space` - Toggle status
- `/` - Search contacts

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Signal Protocol](https://signal.org/docs/) for the cryptographic protocol
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI framework
- [BadgerDB](https://github.com/dgraph-io/badger) for local storage
- The Go community for excellent libraries and tools

## Roadmap

### v0.1.0 - MVP
- [x] Basic architecture design
- [ ] Core encryption implementation
- [ ] Simple TUI interface
- [ ] 1:1 messaging
- [ ] Local message storage

### v0.2.0 - Enhanced
- [ ] Group chats
- [ ] File sharing
- [ ] Improved UI/UX
- [ ] Cross-platform builds

### v0.3.0 - Advanced
- [ ] P2P connections
- [ ] Advanced notifications
- [ ] Plugin system
- [ ] Mobile companion app

---

**SecureChat** - Secure, minimal, developer-focused terminal chat.
