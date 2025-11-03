# SecureChat - Terminal UI Design

## Design Philosophy

### Core Principles
- **Minimal Distraction:** Clean, focused interface
- **Keyboard-First:** All actions accessible via hotkeys
- **Developer-Friendly:** Familiar patterns from terminal tools
- **Accessibility:** Screen reader compatible
- **Performance:** Responsive even on slow terminals

### Visual Hierarchy
1. **Primary:** Active chat messages
2. **Secondary:** Contact list, status indicators
3. **Tertiary:** System messages, timestamps
4. **Minimal:** Borders, decorative elements

## Layout Design

### Main Interface
```
┌─ SecureChat ────────────────────────────────────────────────────┐
│ [Alice] [Bob] [DevTeam] [+]                              [●] Online │
├─────────────────────────────────────────────────────────────────┤
│ Alice                                                    10:30 AM │
│ Hey, are you free for a quick code review?                      │
│                                                                  │
│ You                                                      10:32 AM │
│ Sure! Which PR?                                                  │
│                                                                  │
│ Alice                                                    10:33 AM │
│ The authentication refactor - #247                              │
│ https://github.com/company/repo/pull/247                        │
│                                                                  │
│ You                                                      10:35 AM │
│ Looking at it now. The JWT validation looks good but I have     │
│ a question about the refresh token handling...                  │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│ Type a message... (Ctrl+S to send, Ctrl+Q to quit)             │
└─────────────────────────────────────────────────────────────────┘
```

### Contact List View
```
┌─ Contacts ──────────────────────────────────────────────────────┐
│ Search: [____________]                                    (Esc) │
├─────────────────────────────────────────────────────────────────┤
│ ● Alice Cooper          Last seen: 2 minutes ago               │
│   Software Engineer     "Working on the auth system"           │
│                                                                  │
│ ● Bob Wilson           Last seen: 15 minutes ago               │
│   DevOps Engineer      "Deploying to staging"                  │
│                                                                  │
│ ○ Charlie Davis        Last seen: 2 hours ago                  │
│   Frontend Dev         "In a meeting"                          │
│                                                                  │
│ ○ DevTeam (3)          Last message: 1 hour ago                │
│   Group Chat           Alice: "Meeting at 3pm"                 │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│ [Enter] Open chat  [Space] Toggle status  [Del] Remove contact │
└─────────────────────────────────────────────────────────────────┘
```

### Settings View
```
┌─ Settings ──────────────────────────────────────────────────────┐
│                                                                  │
│ Profile                                                          │
│ ├─ Display Name: [Alice Cooper____________]                     │
│ ├─ Status Message: [Working on auth system_]                    │
│ └─ User ID: alice_cooper_dev (read-only)                        │
│                                                                  │
│ Security                                                         │
│ ├─ Auto-accept keys: [ ] No  [●] Ask  [ ] Yes                  │
│ ├─ Message retention: [30 days ▼]                              │
│ └─ Export keys: [Export...] [Import...]                        │
│                                                                  │
│ Interface                                                        │
│ ├─ Theme: [●] Dark  [ ] Light  [ ] Auto                        │
│ ├─ Notifications: [●] Enabled  [ ] Disabled                    │
│ ├─ Sound alerts: [ ] Enabled  [●] Disabled                     │
│ └─ Timestamp format: [HH:MM ▼]                                 │
│                                                                  │
│ Network                                                          │
│ ├─ Relay servers: [relay1.securechat.dev:8080]                │
│ ├─ P2P connections: [●] Enabled  [ ] Disabled                  │
│ └─ Connection timeout: [30 seconds ▼]                          │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│ [Tab] Next section  [Enter] Edit  [Esc] Back to chat           │
└─────────────────────────────────────────────────────────────────┘
```

## Notification System

### Toast Notifications
```
┌─ New Message ─┐
│ Alice Cooper  │
│ Hey, are you  │
│ free for a... │
│               │
│ [Enter] Reply │
│ [Esc] Dismiss │
└───────────────┘
```

### Status Bar Indicators
```
[●] Online    [◐] Away    [○] Offline    [⚠] Connecting    [✗] Error
[🔒] Encrypted    [🔓] Unencrypted    [⚡] P2P    [🌐] Relay
```

## Keyboard Shortcuts

### Global Shortcuts
| Key | Action |
|-----|--------|
| `Ctrl+N` | New chat |
| `Ctrl+T` | Switch chat tabs |
| `Ctrl+W` | Close current chat |
| `Ctrl+Q` | Quit application |
| `Ctrl+,` | Open settings |
| `Ctrl+/` | Show help |

### Chat View
| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Shift+Enter` | New line |
| `Ctrl+S` | Send message (alternative) |
| `Up/Down` | Navigate message history |
| `Ctrl+L` | Clear chat history |
| `Ctrl+F` | Search messages |

### Contact Management
| Key | Action |
|-----|--------|
| `Ctrl+A` | Add contact |
| `Ctrl+E` | Edit contact |
| `Del` | Remove contact |
| `Space` | Toggle online status |
| `/` | Search contacts |

## Themes

### Dark Theme (Default)
```
Background: #1a1a1a
Foreground: #e0e0e0
Accent: #00d4aa
Border: #404040
Highlight: #2d2d2d
Error: #ff6b6b
Warning: #ffd93d
Success: #6bcf7f
```

### Light Theme
```
Background: #ffffff
Foreground: #2d2d2d
Accent: #0066cc
Border: #d0d0d0
Highlight: #f5f5f5
Error: #dc3545
Warning: #ffc107
Success: #28a745
```

## Responsive Design

### Minimum Terminal Size
- **Width:** 80 characters
- **Height:** 24 lines
- **Graceful degradation:** Hide non-essential elements

### Adaptive Layout
```
Wide Terminal (>120 cols):
┌─ Contacts ─┐ ┌─ Chat ──────────────────┐ ┌─ Info ─┐
│            │ │                         │ │        │
│  Contact   │ │      Messages           │ │ User   │
│   List     │ │                         │ │ Status │
│            │ │                         │ │        │
└────────────┘ └─────────────────────────┘ └────────┘

Narrow Terminal (<80 cols):
┌─ Chat ──────────────────────────────────┐
│              Messages                    │
│                                          │
│ [Contacts] [Settings] [Help]            │
└──────────────────────────────────────────┘
```

## Accessibility Features

### Screen Reader Support
- **ARIA labels:** All interactive elements
- **Focus indicators:** Clear visual focus
- **Semantic markup:** Proper heading hierarchy
- **Alt text:** For status indicators and icons

### Keyboard Navigation
- **Tab order:** Logical navigation flow
- **Focus traps:** Modal dialogs contain focus
- **Skip links:** Jump to main content
- **Shortcuts:** All mouse actions have keyboard equivalents

### Visual Accessibility
- **High contrast:** WCAG AA compliance
- **Font scaling:** Respect terminal font settings
- **Color blind friendly:** Don't rely solely on color
- **Reduced motion:** Minimal animations

## Error Handling UI

### Connection Errors
```
┌─ Connection Error ─────────────────────────────────────────────┐
│ ⚠ Unable to connect to relay server                           │
│                                                                │
│ • Check your internet connection                               │
│ • Verify server address in settings                           │
│ • Try connecting to a different relay                         │
│                                                                │
│ [Retry] [Settings] [Use P2P Only]                            │
└────────────────────────────────────────────────────────────────┘
```

### Encryption Errors
```
┌─ Security Warning ─────────────────────────────────────────────┐
│ 🔒 Unable to verify Alice's identity                          │
│                                                                │
│ Their security key has changed. This could mean:              │
│ • They reinstalled SecureChat                                 │
│ • Someone is intercepting your messages                       │
│                                                                │
│ Safety Number: 12345 67890 12345 67890 12345 67890           │
│                                                                │
│ [Verify] [Accept] [Block Contact]                            │
└────────────────────────────────────────────────────────────────┘
```

## Animation & Feedback

### Subtle Animations
- **Message arrival:** Gentle slide-in effect
- **Typing indicators:** Pulsing dots
- **Connection status:** Smooth color transitions
- **Focus changes:** Soft highlight transitions

### Loading States
```
Connecting...  [●○○]
Sending...     [●●○]  
Encrypting...  [●●●]
```

### Progress Indicators
```
File Transfer: alice_photo.jpg
[████████████████████████████████████████] 100% (2.4 MB)
Speed: 1.2 MB/s  ETA: Complete
```
