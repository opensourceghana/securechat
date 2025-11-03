package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shelemiah/secure_chat/internal/config"
)

// HelpView represents the help interface
type HelpView struct {
	config *config.Config
	theme  *Theme
	width  int
	height int
	
	// Help state
	selectedSection int
	scrollOffset    int
	
	// Help sections
	sections []HelpSection
}

// HelpSection represents a group of related help topics
type HelpSection struct {
	Name    string
	Content []string
}

// NewHelpView creates a new help view
func NewHelpView(cfg *config.Config, theme *Theme) *HelpView {
	view := &HelpView{
		config: cfg,
		theme:  theme,
	}
	
	view.sections = buildHelpSections()
	
	return view
}

// Init implements tea.Model
func (h *HelpView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (h *HelpView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height - 2 // Account for status bar
		
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if h.scrollOffset > 0 {
				h.scrollOffset--
			}
			
		case "down", "j":
			maxScroll := h.getMaxScroll()
			if h.scrollOffset < maxScroll {
				h.scrollOffset++
			}
			
		case "left", "h":
			if h.selectedSection > 0 {
				h.selectedSection--
				h.scrollOffset = 0
			}
			
		case "right", "l":
			if h.selectedSection < len(h.sections)-1 {
				h.selectedSection++
				h.scrollOffset = 0
			}
			
		case "tab":
			h.selectedSection = (h.selectedSection + 1) % len(h.sections)
			h.scrollOffset = 0
			
		case "home":
			h.scrollOffset = 0
			
		case "end":
			h.scrollOffset = h.getMaxScroll()
		}
	}
	
	return h, nil
}

// View implements tea.Model
func (h *HelpView) View() string {
	if h.width == 0 || h.height == 0 {
		return "Loading help..."
	}
	
	// Create main layout
	header := h.renderHeader()
	content := h.renderContent()
	footer := h.renderFooter()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

// renderHeader renders the help view header
func (h *HelpView) renderHeader() string {
	style := lipgloss.NewStyle().
		Background(h.theme.Primary).
		Foreground(h.theme.Background).
		Padding(0, 1).
		Width(h.width)
	
	title := "Help & Keyboard Shortcuts"
	
	// Section tabs
	var tabs []string
	for i, section := range h.sections {
		var tabStyle lipgloss.Style
		if i == h.selectedSection {
			tabStyle = lipgloss.NewStyle().
				Background(h.theme.Background).
				Foreground(h.theme.Primary).
				Padding(0, 1).
				Bold(true)
		} else {
			tabStyle = lipgloss.NewStyle().
				Background(h.theme.Secondary).
				Foreground(h.theme.Background).
				Padding(0, 1)
		}
		tabs = append(tabs, tabStyle.Render(section.Name))
	}
	
	tabsContent := strings.Join(tabs, " ")
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.Render(title),
		tabsContent,
	)
}

// renderContent renders the help content
func (h *HelpView) renderContent() string {
	contentHeight := h.height - 4 // Account for header and footer
	
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(h.theme.Border).
		Width(h.width).
		Height(contentHeight).
		Padding(1)
	
	section := h.sections[h.selectedSection]
	visibleContent := h.getVisibleContent(section.Content)
	
	content := strings.Join(visibleContent, "\n")
	
	return style.Render(content)
}

// renderFooter renders the help view footer
func (h *HelpView) renderFooter() string {
	style := lipgloss.NewStyle().
		Background(h.theme.Secondary).
		Foreground(h.theme.Background).
		Padding(0, 1).
		Width(h.width)
	
	shortcuts := "[←→] Switch sections  [↑↓] Scroll  [Tab] Next section  [Esc] Back to chat"
	
	return style.Render(shortcuts)
}

// getVisibleContent returns the content lines that should be visible
func (h *HelpView) getVisibleContent(content []string) []string {
	contentHeight := h.height - 4 - 2 // Account for header, footer, and padding
	
	if len(content) <= contentHeight {
		return content
	}
	
	start := h.scrollOffset
	end := start + contentHeight
	
	if end > len(content) {
		end = len(content)
	}
	
	return content[start:end]
}

// getMaxScroll returns the maximum scroll offset
func (h *HelpView) getMaxScroll() int {
	contentHeight := h.height - 4 - 2
	section := h.sections[h.selectedSection]
	
	maxScroll := len(section.Content) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	
	return maxScroll
}

// buildHelpSections creates the help content sections
func buildHelpSections() []HelpSection {
	return []HelpSection{
		{
			Name: "General",
			Content: []string{
				"Welcome to SecureChat!",
				"",
				"SecureChat is a secure, terminal-based chat application designed for",
				"developers who want distraction-free communication while staying in",
				"their terminal workflow.",
				"",
				"Key Features:",
				"• End-to-end encryption using the Signal Protocol",
				"• Terminal-native interface with minimal distractions",
				"• Peer-to-peer connections with relay fallback",
				"• Cross-platform support (Linux, macOS, Windows)",
				"• Keyboard-driven interface with vim-like bindings",
				"",
				"Getting Started:",
				"1. Add contacts using Ctrl+A in the contacts view",
				"2. Start chatting by selecting a contact and pressing Enter",
				"3. Use F1/F2 to switch between chat and contacts views",
				"4. Configure settings with Ctrl+,",
				"",
				"Security:",
				"All messages are encrypted end-to-end. Even relay servers cannot",
				"read your messages. Always verify contact identities using safety",
				"numbers for maximum security.",
			},
		},
		{
			Name: "Keyboard Shortcuts",
			Content: []string{
				"Global Shortcuts:",
				"",
				"Ctrl+Q          Quit SecureChat",
				"Ctrl+C          Quit SecureChat (alternative)",
				"Ctrl+,          Open settings",
				"Ctrl+/          Show this help",
				"F1              Switch to chat view",
				"F2              Switch to contacts view",
				"Esc             Return to chat from other views",
				"",
				"Chat View:",
				"",
				"Enter           Send message",
				"Shift+Enter     New line in message",
				"Ctrl+S          Send message (alternative)",
				"Ctrl+L          Clear chat history",
				"Ctrl+F          Search messages",
				"Ctrl+N          New chat",
				"Ctrl+T          Switch between chat tabs",
				"Ctrl+W          Close current chat",
				"↑/↓             Navigate message history",
				"",
				"Contacts View:",
				"",
				"Enter           Open chat with selected contact",
				"Ctrl+A          Add new contact",
				"Ctrl+E          Edit selected contact",
				"Delete/X        Remove selected contact",
				"Space           Toggle contact status",
				"/               Search contacts",
				"↑/↓ or J/K      Navigate contact list",
				"",
				"Settings View:",
				"",
				"Tab             Next settings section",
				"Enter           Edit selected setting",
				"Space           Toggle boolean settings",
				"↑/↓ or J/K      Navigate settings",
				"←/→ or H/L      Switch settings sections",
			},
		},
		{
			Name: "Security",
			Content: []string{
				"SecureChat Security Overview:",
				"",
				"End-to-End Encryption:",
				"• All messages are encrypted using the Signal Protocol",
				"• Perfect Forward Secrecy: past messages remain secure",
				"• Post-Compromise Security: recovery from key compromise",
				"• ChaCha20-Poly1305 for message encryption",
				"• X25519 for key exchange, Ed25519 for signatures",
				"",
				"Identity Verification:",
				"• Each contact has a unique safety number",
				"• Compare safety numbers out-of-band (voice call, in-person)",
				"• Verified contacts show a checkmark (✓)",
				"• Always verify important contacts manually",
				"",
				"Safety Numbers:",
				"Safety numbers are 60-digit codes that uniquely identify",
				"the cryptographic relationship between you and a contact.",
				"",
				"Example: 12345 67890 12345 67890 12345 67890",
				"",
				"To verify a contact:",
				"1. Go to contacts view (F2)",
				"2. Select the contact and press Enter",
				"3. Compare the safety number with your contact",
				"4. Mark as verified if numbers match",
				"",
				"Network Security:",
				"• TLS 1.3 for relay server connections",
				"• Certificate pinning prevents MITM attacks",
				"• Optional P2P connections bypass servers entirely",
				"• Traffic analysis protection with message padding",
				"",
				"Local Security:",
				"• Messages encrypted before local storage",
				"• Cryptographic keys stored in OS keychain",
				"• Secure deletion of old message keys",
				"• Configurable message retention periods",
				"",
				"Best Practices:",
				"• Always verify contact identities",
				"• Use P2P connections when possible",
				"• Regularly update SecureChat",
				"• Use strong device passwords/PINs",
				"• Be cautious with relay server selection",
			},
		},
		{
			Name: "Configuration",
			Content: []string{
				"Configuration File:",
				"",
				"SecureChat looks for configuration in these locations:",
				"1. ~/.config/securechat/config.yaml",
				"2. ~/.securechat.yaml",
				"3. ./config.yaml",
				"",
				"Example configuration:",
				"",
				"user:",
				"  display_name: \"Your Name\"",
				"  status_message: \"Available\"",
				"",
				"network:",
				"  relay_servers:",
				"    - \"relay1.securechat.dev:8080\"",
				"    - \"relay2.securechat.dev:8080\"",
				"  p2p_enabled: true",
				"  connection_timeout: \"30s\"",
				"",
				"ui:",
				"  theme: \"dark\"",
				"  notifications: true",
				"  sound_enabled: false",
				"  timestamp_format: \"15:04\"",
				"",
				"security:",
				"  auto_accept_keys: false",
				"  message_retention_days: 30",
				"  require_verification: true",
				"",
				"Themes:",
				"• dark: Dark theme (default)",
				"• light: Light theme",
				"• auto: Follow system theme",
				"",
				"Relay Servers:",
				"Relay servers help route messages when P2P connections",
				"aren't possible. They cannot read your encrypted messages.",
				"",
				"You can:",
				"• Use public relay servers",
				"• Run your own relay server",
				"• Use P2P-only mode (no relays)",
				"",
				"Message Retention:",
				"Configure how long to keep message history:",
				"• 1 day to 1 year",
				"• Forever (not recommended for security)",
				"• Messages are securely deleted after retention period",
			},
		},
		{
			Name: "Troubleshooting",
			Content: []string{
				"Common Issues and Solutions:",
				"",
				"Connection Problems:",
				"• Check internet connection",
				"• Verify relay server addresses in settings",
				"• Try enabling/disabling P2P connections",
				"• Check firewall settings for P2P mode",
				"",
				"Message Delivery Issues:",
				"• Ensure both users are online",
				"• Check if contact's keys have changed",
				"• Verify network connectivity",
				"• Try switching relay servers",
				"",
				"Encryption Errors:",
				"• Contact's identity key may have changed",
				"• Verify safety numbers with contact",
				"• Clear and re-establish session if needed",
				"• Check for app updates",
				"",
				"Performance Issues:",
				"• Clear old message history (Ctrl+L)",
				"• Reduce message retention period",
				"• Check available disk space",
				"• Restart SecureChat",
				"",
				"UI Problems:",
				"• Resize terminal window",
				"• Try different theme in settings",
				"• Check terminal color support",
				"• Update terminal emulator",
				"",
				"Getting Help:",
				"• Check GitHub issues: github.com/shelemiah/secure_chat",
				"• Read documentation: docs.securechat.dev",
				"• Join community chat (coming soon)",
				"",
				"Reporting Bugs:",
				"• Include SecureChat version (securechat --version)",
				"• Describe steps to reproduce",
				"• Include relevant log output",
				"• Mention your operating system",
				"",
				"Security Issues:",
				"• Report to: security@securechat.dev",
				"• Use PGP encryption if possible",
				"• Allow 24-48 hours for response",
				"• Coordinate responsible disclosure",
			},
		},
	}
}
