# SecureChat - Technical Architecture

## System Overview

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

## Component Architecture

### Client Components

#### 1. Terminal User Interface (TUI)
- **Framework:** Bubble Tea (modern, composable)
- **Responsibilities:**
  - Message display and input
  - Contact management
  - Settings configuration
  - Toast notifications
- **Key Features:**
  - Vim-like keybindings
  - Multiple views (chat, contacts, settings)
  - Responsive layout
  - Theme support

#### 2. Cryptography Module
- **Library:** NaCl/libsodium (via Go bindings)
- **Responsibilities:**
  - End-to-end encryption
  - Key generation and management
  - Digital signatures
  - Perfect forward secrecy
- **Algorithms:**
  - Encryption: ChaCha20-Poly1305
  - Key exchange: X25519
  - Signatures: Ed25519
  - Hashing: BLAKE2b

#### 3. Network Layer
- **Protocol:** Custom over TCP/WebSocket
- **Responsibilities:**
  - Connection management
  - Message routing
  - NAT traversal (STUN/TURN)
  - Reconnection logic
- **Features:**
  - Connection pooling
  - Automatic retry
  - Bandwidth optimization

#### 4. Storage Layer
- **Database:** BadgerDB (embedded key-value store)
- **Responsibilities:**
  - Message history
  - Contact information
  - Cryptographic keys
  - Configuration settings
- **Structure:**
```
/contacts/{user_id} -> ContactInfo
/messages/{chat_id}/{timestamp} -> EncryptedMessage
/keys/{key_id} -> CryptoKey
/config/{setting} -> Value
```

### Server Components (Optional Relay)

#### 1. Message Router
- **Responsibilities:**
  - Route messages between clients
  - Handle offline message queuing
  - Manage connection state
- **Features:**
  - Load balancing
  - Rate limiting
  - Spam protection

#### 2. Discovery Service
- **Responsibilities:**
  - User registration
  - Contact discovery
  - Presence management
- **Privacy:**
  - Minimal metadata storage
  - No message content access
  - Regular data purging

## Message Protocol

### Message Format
```json
{
  "version": "1.0",
  "type": "chat|system|presence",
  "from": "user_id",
  "to": "user_id|group_id",
  "timestamp": 1699000000,
  "encrypted_payload": "base64_encrypted_data",
  "signature": "base64_signature"
}
```

### Encryption Flow
1. **Key Exchange:** X25519 ECDH
2. **Derive Keys:** HKDF with message counter
3. **Encrypt:** ChaCha20-Poly1305 AEAD
4. **Sign:** Ed25519 signature
5. **Transmit:** Base64 encoded payload

## Security Model

### Threat Model
- **Adversaries:**
  - Network eavesdroppers
  - Malicious relay servers
  - Compromised endpoints
- **Protections:**
  - End-to-end encryption
  - Forward secrecy
  - Authentication
  - Metadata minimization

### Key Management
```
Identity Key (Long-term)
├── Signing Key (Ed25519)
└── Exchange Key (X25519)

Session Keys (Ephemeral)
├── Root Key
├── Chain Key
└── Message Keys
```

## Data Flow

### Message Sending
```
User Input → TUI → Crypto.Encrypt() → Network.Send() → [Relay] → Recipient
```

### Message Receiving
```
Network.Receive() → Crypto.Decrypt() → Storage.Save() → TUI.Display() → Notification
```

## Configuration

### Client Configuration
```yaml
# ~/.config/securechat/config.yaml
user:
  id: "user_12345"
  display_name: "Alice"
  
network:
  relay_servers:
    - "relay1.securechat.dev:8080"
    - "relay2.securechat.dev:8080"
  p2p_enabled: true
  
ui:
  theme: "dark"
  notifications: true
  sound_enabled: false
  
security:
  auto_accept_keys: false
  message_retention_days: 30
```

## Deployment Options

### 1. Standalone Binary
- Single executable
- Embedded assets
- Cross-platform builds

### 2. Package Managers
- Homebrew (macOS)
- APT/RPM (Linux)
- Chocolatey (Windows)
- Go modules

### 3. Container
- Docker image
- Kubernetes deployment
- Easy server hosting

## Performance Considerations

### Memory Usage
- Target: < 50MB RAM
- Efficient message storage
- Connection pooling
- Garbage collection optimization

### Network Efficiency
- Message compression
- Connection reuse
- Batch operations
- Adaptive quality

### Startup Time
- Target: < 500ms cold start
- Lazy loading
- Minimal dependencies
- Optimized binary size

## Scalability

### Client Scalability
- Support 100+ contacts
- 10,000+ message history
- Multiple concurrent chats

### Server Scalability
- Horizontal scaling
- Database sharding
- CDN for static assets
- Microservice architecture
