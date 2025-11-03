# SecureChat - Network Protocol Specification

## Protocol Overview

### Design Goals
- **Simple:** Easy to implement and debug
- **Secure:** End-to-end encryption by default
- **Efficient:** Minimal bandwidth usage
- **Extensible:** Support for future features
- **Resilient:** Handle network failures gracefully

### Transport Layer
- **Primary:** WebSocket over TLS 1.3
- **Fallback:** TCP with custom framing
- **P2P:** Direct connections when possible
- **Relay:** Server-mediated for NAT traversal

## Message Format

### Wire Protocol
```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│   Magic     │   Version   │   Length    │   Type      │
│  (4 bytes)  │  (2 bytes)  │  (4 bytes)  │  (2 bytes)  │
├─────────────┼─────────────┼─────────────┼─────────────┤
│                    Payload                            │
│                 (Length bytes)                        │
└───────────────────────────────────────────────────────┘

Magic: 0x53434854 ("SCHT")
Version: 0x0001 (v1.0)
Length: Payload size in bytes (big-endian)
Type: Message type identifier
```

### JSON Payload Structure
```json
{
  "id": "msg_1699000000_abc123",
  "type": "chat|presence|system|ack",
  "from": "user_alice_123",
  "to": "user_bob_456",
  "timestamp": 1699000000,
  "encrypted": true,
  "payload": "base64_encrypted_data",
  "signature": "base64_signature",
  "metadata": {
    "reply_to": "msg_id",
    "thread_id": "thread_id",
    "expires_at": 1699086400
  }
}
```

## Message Types

### 1. Authentication Messages

#### Client Hello
```json
{
  "type": "client_hello",
  "version": "1.0",
  "user_id": "alice_123",
  "public_key": "base64_ed25519_key",
  "challenge": "random_32_bytes",
  "capabilities": ["e2e_encryption", "file_transfer", "groups"]
}
```

#### Server Hello
```json
{
  "type": "server_hello",
  "version": "1.0",
  "server_id": "relay_server_1",
  "challenge_response": "signed_challenge",
  "session_id": "session_abc123",
  "capabilities": ["message_relay", "offline_storage", "push_notifications"]
}
```

### 2. Chat Messages

#### Text Message
```json
{
  "type": "chat",
  "content_type": "text/plain",
  "encrypted_payload": {
    "text": "Hello, world!",
    "formatting": null
  }
}
```

#### Rich Message
```json
{
  "type": "chat", 
  "content_type": "text/markdown",
  "encrypted_payload": {
    "text": "Check out this **important** [link](https://example.com)",
    "formatting": "markdown",
    "attachments": [
      {
        "type": "link_preview",
        "url": "https://example.com",
        "title": "Example Site",
        "description": "An example website"
      }
    ]
  }
}
```

### 3. Presence Messages

#### Status Update
```json
{
  "type": "presence",
  "status": "online|away|busy|offline",
  "status_message": "Working on SecureChat",
  "last_seen": 1699000000
}
```

#### Typing Indicator
```json
{
  "type": "typing",
  "chat_id": "chat_alice_bob",
  "typing": true,
  "expires_at": 1699000005
}
```

### 4. System Messages

#### Key Exchange
```json
{
  "type": "key_exchange",
  "key_type": "prekey|identity|ephemeral",
  "public_key": "base64_x25519_key",
  "key_id": "key_123",
  "signature": "base64_signature"
}
```

#### Error Response
```json
{
  "type": "error",
  "error_code": "INVALID_SIGNATURE",
  "error_message": "Message signature verification failed",
  "reference_id": "msg_1699000000_abc123"
}
```

## Connection Management

### Connection States
```
DISCONNECTED → CONNECTING → AUTHENTICATING → CONNECTED
     ↑              ↓              ↓            ↓
     └──────────────┴──────────────┴────────────┘
                    (on error/timeout)
```

### Heartbeat Protocol
```json
// Every 30 seconds
{
  "type": "ping",
  "timestamp": 1699000000,
  "sequence": 12345
}

// Response
{
  "type": "pong", 
  "timestamp": 1699000001,
  "sequence": 12345
}
```

### Reconnection Logic
1. **Exponential Backoff:** 1s, 2s, 4s, 8s, 16s, 30s (max)
2. **Jitter:** ±25% random variation
3. **Circuit Breaker:** Stop after 10 consecutive failures
4. **Health Check:** Ping before declaring connection healthy

## Peer-to-Peer Protocol

### Discovery
```json
{
  "type": "p2p_offer",
  "from": "alice_123",
  "to": "bob_456", 
  "sdp_offer": "webrtc_session_description",
  "ice_candidates": ["candidate_1", "candidate_2"]
}
```

### NAT Traversal
- **STUN Servers:** Public STUN servers for NAT detection
- **TURN Servers:** Relay servers as last resort
- **ICE:** Interactive Connectivity Establishment
- **Hole Punching:** UDP hole punching for direct connections

### P2P Message Format
```json
{
  "type": "p2p_message",
  "session_id": "p2p_session_abc",
  "sequence": 1,
  "encrypted_payload": "base64_data",
  "signature": "base64_signature"
}
```

## Group Chat Protocol

### Group Management
```json
{
  "type": "group_create",
  "group_id": "group_abc123",
  "group_name": "DevTeam",
  "members": ["alice_123", "bob_456", "charlie_789"],
  "admin": "alice_123",
  "group_key": "encrypted_group_key"
}
```

### Member Operations
```json
{
  "type": "group_member_add",
  "group_id": "group_abc123", 
  "new_member": "dave_101",
  "invited_by": "alice_123",
  "group_key": "encrypted_for_dave"
}
```

### Message Distribution
- **Fan-out:** Send to each member individually
- **Key Rotation:** New group key when members change
- **Delivery Receipts:** Track message delivery per member

## File Transfer Protocol

### File Offer
```json
{
  "type": "file_offer",
  "file_id": "file_abc123",
  "filename": "document.pdf",
  "file_size": 1048576,
  "mime_type": "application/pdf",
  "checksum": "sha256_hash",
  "encryption_key": "base64_key"
}
```

### Transfer Chunks
```json
{
  "type": "file_chunk",
  "file_id": "file_abc123",
  "chunk_index": 0,
  "total_chunks": 100,
  "encrypted_data": "base64_chunk_data"
}
```

### Progress Tracking
```json
{
  "type": "file_progress",
  "file_id": "file_abc123", 
  "bytes_received": 524288,
  "total_bytes": 1048576,
  "status": "downloading|complete|error"
}
```

## Error Handling

### Error Codes
| Code | Description |
|------|-------------|
| 1000 | Invalid message format |
| 1001 | Authentication failed |
| 1002 | Invalid signature |
| 1003 | Encryption error |
| 1004 | Rate limit exceeded |
| 1005 | User not found |
| 1006 | Permission denied |
| 2000 | Network timeout |
| 2001 | Connection lost |
| 2002 | Server unavailable |

### Retry Strategy
```json
{
  "type": "retry_policy",
  "max_retries": 3,
  "backoff_multiplier": 2.0,
  "initial_delay_ms": 1000,
  "max_delay_ms": 30000,
  "jitter_factor": 0.1
}
```

## Rate Limiting

### Client Limits
- **Messages:** 100 per minute per user
- **Connections:** 5 concurrent per user
- **File Transfers:** 10 MB per minute per user
- **Group Operations:** 10 per minute per user

### Server Limits
- **Global Messages:** 10,000 per second
- **Per-IP Connections:** 100 concurrent
- **Bandwidth:** 1 GB per minute per server
- **Storage:** 1 GB per user

## Security Considerations

### Message Ordering
- **Sequence Numbers:** Prevent replay attacks
- **Timestamps:** Detect delayed messages
- **Nonces:** Ensure message uniqueness

### Traffic Analysis Protection
- **Message Padding:** Random padding to fixed sizes
- **Dummy Traffic:** Send fake messages periodically
- **Timing Randomization:** Add random delays

### Forward Secrecy
- **Key Rotation:** Rotate keys every 1000 messages
- **Key Deletion:** Securely delete old keys
- **Session Keys:** Unique keys per session

## Implementation Notes

### Go Libraries
```go
// WebSocket
"github.com/gorilla/websocket"

// Cryptography
"golang.org/x/crypto/nacl/box"
"golang.org/x/crypto/ed25519"

// Networking
"net"
"crypto/tls"

// JSON Processing
"encoding/json"
```

### Performance Optimizations
- **Connection Pooling:** Reuse connections
- **Message Batching:** Send multiple messages together
- **Compression:** Optional gzip compression
- **Binary Protocol:** Consider protobuf for efficiency

### Testing Strategy
- **Unit Tests:** Protocol message parsing
- **Integration Tests:** End-to-end message flow
- **Load Tests:** High message volume scenarios
- **Security Tests:** Fuzzing and penetration testing
