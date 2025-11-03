package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shelemiah/secure_chat/internal/config"
	"github.com/shelemiah/secure_chat/internal/models"
)

// ChatView represents the main chat interface
type ChatView struct {
	config   *config.Config
	theme    *Theme
	width    int
	height   int
	
	// Chat state
	messages    []models.Message
	currentChat string
	input       string
	cursor      int
	
	// UI state
	scrollOffset int
	typing       bool
}

// NewChatView creates a new chat view
func NewChatView(cfg *config.Config, theme *Theme) *ChatView {
	return &ChatView{
		config:   cfg,
		theme:    theme,
		messages: []models.Message{},
	}
}

// Init implements tea.Model
func (c *ChatView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (c *ChatView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height - 2 // Account for status bar
		
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if c.input != "" {
				// Send message
				newMsg := models.NewMessage(
					models.MessageTypeChat,
					c.config.User.ID,
					c.currentChat,
					c.input,
				)
				c.messages = append(c.messages, *newMsg)
				c.input = ""
				c.cursor = 0
				c.scrollToBottom()
			}
			
		case "backspace":
			if c.cursor > 0 {
				c.input = c.input[:c.cursor-1] + c.input[c.cursor:]
				c.cursor--
			}
			
		case "left":
			if c.cursor > 0 {
				c.cursor--
			}
			
		case "right":
			if c.cursor < len(c.input) {
				c.cursor++
			}
			
		case "up":
			if c.scrollOffset > 0 {
				c.scrollOffset--
			}
			
		case "down":
			maxScroll := len(c.messages) - c.getMessageAreaHeight()
			if c.scrollOffset < maxScroll {
				c.scrollOffset++
			}
			
		case "ctrl+l":
			c.messages = []models.Message{}
			c.scrollOffset = 0
			
		default:
			// Handle regular character input
			if len(msg.String()) == 1 {
				c.input = c.input[:c.cursor] + msg.String() + c.input[c.cursor:]
				c.cursor++
			}
		}
	}
	
	return c, nil
}

// View implements tea.Model
func (c *ChatView) View() string {
	if c.width == 0 || c.height == 0 {
		return "Loading chat..."
	}
	
	// Create main layout
	header := c.renderHeader()
	messages := c.renderMessages()
	input := c.renderInput()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		messages,
		input,
	)
}

// renderHeader renders the chat header with contact info
func (c *ChatView) renderHeader() string {
	style := lipgloss.NewStyle().
		Background(c.theme.Primary).
		Foreground(c.theme.Background).
		Padding(0, 1).
		Width(c.width)
	
	title := "SecureChat"
	if c.currentChat != "" {
		title = fmt.Sprintf("Chat with %s", c.currentChat)
	}
	
	status := "● Online"
	
	// Center title, right-align status
	padding := c.width - len(title) - len(status) - 4 // Account for padding
	if padding < 0 {
		padding = 0
	}
	
	content := fmt.Sprintf("%s%s%s", 
		title,
		strings.Repeat(" ", padding),
		status,
	)
	
	return style.Render(content)
}

// renderMessages renders the message area
func (c *ChatView) renderMessages() string {
	messageHeight := c.getMessageAreaHeight()
	
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(c.theme.Border).
		Width(c.width).
		Height(messageHeight).
		Padding(1)
	
	if len(c.messages) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(c.theme.Secondary).
			Italic(true).
			Align(lipgloss.Center).
			Width(c.width - 4).
			Height(messageHeight - 2)
		
		return style.Render(
			emptyStyle.Render("No messages yet. Start typing to send a message!"),
		)
	}
	
	// Render visible messages
	var messageLines []string
	visibleMessages := c.getVisibleMessages()
	
	for _, msg := range visibleMessages {
		messageLines = append(messageLines, c.formatMessage(msg))
	}
	
	content := strings.Join(messageLines, "\n")
	
	return style.Render(content)
}

// renderInput renders the message input area
func (c *ChatView) renderInput() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(c.theme.Border).
		Width(c.width).
		Padding(0, 1)
	
	prompt := "Type a message... "
	
	// Show cursor
	input := c.input
	if c.cursor < len(input) {
		input = input[:c.cursor] + "│" + input[c.cursor:]
	} else {
		input += "│"
	}
	
	content := prompt + input
	
	// Add help text
	help := lipgloss.NewStyle().
		Foreground(c.theme.Secondary).
		Render("(Enter to send, Ctrl+L to clear, ↑↓ to scroll)")
	
	return style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			help,
		),
	)
}

// formatMessage formats a single message for display
func (c *ChatView) formatMessage(msg models.Message) string {
	timeStr := msg.Timestamp.Format(c.config.UI.TimestampFormat)
	
	var senderStyle lipgloss.Style
	if msg.IsFromUser(c.config.User.ID) {
		senderStyle = lipgloss.NewStyle().
			Foreground(c.theme.Primary).
			Bold(true)
	} else {
		senderStyle = lipgloss.NewStyle().
			Foreground(c.theme.Success).
			Bold(true)
	}
	
	timeStyle := lipgloss.NewStyle().
		Foreground(c.theme.Secondary)
	
	contentStyle := lipgloss.NewStyle().
		Foreground(c.theme.Foreground)
	
	sender := "You"
	if !msg.IsFromUser(c.config.User.ID) {
		sender = msg.From // In real app, this would be display name
	}
	
	return fmt.Sprintf("%s %s\n%s",
		senderStyle.Render(sender),
		timeStyle.Render(timeStr),
		contentStyle.Render(msg.Content),
	)
}

// getVisibleMessages returns messages that should be visible in the current scroll position
func (c *ChatView) getVisibleMessages() []models.Message {
	if len(c.messages) == 0 {
		return []models.Message{}
	}
	
	messageHeight := c.getMessageAreaHeight()
	
	// Calculate how many messages can fit
	// Each message takes approximately 2-3 lines
	maxMessages := messageHeight / 3
	
	start := c.scrollOffset
	end := start + maxMessages
	
	if end > len(c.messages) {
		end = len(c.messages)
	}
	
	if start >= len(c.messages) {
		start = len(c.messages) - 1
	}
	
	return c.messages[start:end]
}

// getMessageAreaHeight returns the height available for messages
func (c *ChatView) getMessageAreaHeight() int {
	// Total height minus header (1) and input area (3) and borders
	return c.height - 1 - 3 - 2
}

// scrollToBottom scrolls to show the latest messages
func (c *ChatView) scrollToBottom() {
	messageHeight := c.getMessageAreaHeight()
	maxMessages := messageHeight / 3
	
	if len(c.messages) > maxMessages {
		c.scrollOffset = len(c.messages) - maxMessages
	} else {
		c.scrollOffset = 0
	}
}
