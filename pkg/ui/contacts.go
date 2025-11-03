package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opensourceghana/securechat/internal/config"
	"github.com/opensourceghana/securechat/internal/models"
)

// ContactsView represents the contacts management interface
type ContactsView struct {
	config   *config.Config
	theme    *Theme
	width    int
	height   int
	
	// Contact state
	contacts     []models.Contact
	selectedIdx  int
	searchQuery  string
	searchActive bool
	
	// UI state
	scrollOffset int
}

// NewContactsView creates a new contacts view
func NewContactsView(cfg *config.Config, theme *Theme) *ContactsView {
	return &ContactsView{
		config:   cfg,
		theme:    theme,
		contacts: generateSampleContacts(), // TODO: Load from storage
	}
}

// Init implements tea.Model
func (c *ContactsView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (c *ContactsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height - 2 // Account for status bar
		
	case tea.KeyMsg:
		if c.searchActive {
			return c.handleSearchInput(msg)
		}
		
		switch msg.String() {
		case "up", "k":
			if c.selectedIdx > 0 {
				c.selectedIdx--
				c.adjustScroll()
			}
			
		case "down", "j":
			if c.selectedIdx < len(c.contacts)-1 {
				c.selectedIdx++
				c.adjustScroll()
			}
			
		case "enter":
			if len(c.contacts) > 0 {
				// TODO: Open chat with selected contact
				return c, nil
			}
			
		case "/":
			c.searchActive = true
			c.searchQuery = ""
			
		case "ctrl+a":
			// TODO: Add new contact
			return c, nil
			
		case "ctrl+e":
			// TODO: Edit selected contact
			return c, nil
			
		case "delete", "x":
			// TODO: Remove selected contact
			return c, nil
			
		case "space":
			// TODO: Toggle contact status
			return c, nil
		}
	}
	
	return c, nil
}

// View implements tea.Model
func (c *ContactsView) View() string {
	if c.width == 0 || c.height == 0 {
		return "Loading contacts..."
	}
	
	// Create main layout
	header := c.renderHeader()
	contactList := c.renderContactList()
	footer := c.renderFooter()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		contactList,
		footer,
	)
}

// handleSearchInput handles keyboard input during search
func (c *ContactsView) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		c.searchActive = false
		c.searchQuery = ""
		
	case "enter":
		c.searchActive = false
		// TODO: Filter contacts based on search query
		
	case "backspace":
		if len(c.searchQuery) > 0 {
			c.searchQuery = c.searchQuery[:len(c.searchQuery)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			c.searchQuery += msg.String()
		}
	}
	
	return c, nil
}

// renderHeader renders the contacts view header
func (c *ContactsView) renderHeader() string {
	style := lipgloss.NewStyle().
		Background(c.theme.Primary).
		Foreground(c.theme.Background).
		Padding(0, 1).
		Width(c.width)
	
	title := "Contacts"
	count := fmt.Sprintf("(%d)", len(c.contacts))
	
	// Search bar
	searchStyle := lipgloss.NewStyle().
		Background(c.theme.Background).
		Foreground(c.theme.Foreground).
		Padding(0, 1).
		Margin(0, 1)
	
	var searchBar string
	if c.searchActive {
		searchBar = searchStyle.Render(fmt.Sprintf("Search: %s│", c.searchQuery))
	} else {
		searchBar = searchStyle.Render("Search: [Press / to search]")
	}
	
	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title+" "+count,
		strings.Repeat(" ", c.width-len(title)-len(count)-len(searchBar)-4),
		"(Esc)",
	)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.Render(headerContent),
		searchBar,
	)
}

// renderContactList renders the list of contacts
func (c *ContactsView) renderContactList() string {
	listHeight := c.height - 4 // Account for header and footer
	
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(c.theme.Border).
		Width(c.width).
		Height(listHeight).
		Padding(1)
	
	if len(c.contacts) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(c.theme.Secondary).
			Italic(true).
			Align(lipgloss.Center).
			Width(c.width - 4).
			Height(listHeight - 2)
		
		return style.Render(
			emptyStyle.Render("No contacts yet. Press Ctrl+A to add a contact."),
		)
	}
	
	// Render visible contacts
	var contactLines []string
	visibleContacts := c.getVisibleContacts()
	
	for i, contact := range visibleContacts {
		actualIdx := c.scrollOffset + i
		isSelected := actualIdx == c.selectedIdx
		
		contactLine := c.formatContact(contact, isSelected)
		contactLines = append(contactLines, contactLine)
	}
	
	content := strings.Join(contactLines, "\n\n")
	
	return style.Render(content)
}

// renderFooter renders the contacts view footer with shortcuts
func (c *ContactsView) renderFooter() string {
	style := lipgloss.NewStyle().
		Background(c.theme.Secondary).
		Foreground(c.theme.Background).
		Padding(0, 1).
		Width(c.width)
	
	shortcuts := "[Enter] Open chat  [Space] Toggle status  [Ctrl+A] Add  [Ctrl+E] Edit  [Del] Remove"
	
	return style.Render(shortcuts)
}

// formatContact formats a single contact for display
func (c *ContactsView) formatContact(contact models.Contact, isSelected bool) string {
	var style lipgloss.Style
	if isSelected {
		style = lipgloss.NewStyle().
			Background(c.theme.Highlight).
			Foreground(c.theme.Foreground).
			Padding(0, 1)
	} else {
		style = lipgloss.NewStyle().
			Foreground(c.theme.Foreground)
	}
	
	// Status indicator
	var statusIcon string
	var statusColor lipgloss.Color
	switch contact.Status {
	case models.UserStatusOnline:
		statusIcon = "●"
		statusColor = c.theme.Success
	case models.UserStatusAway:
		statusIcon = "◐"
		statusColor = c.theme.Warning
	case models.UserStatusBusy:
		statusIcon = "◐"
		statusColor = c.theme.Error
	default:
		statusIcon = "○"
		statusColor = c.theme.Secondary
	}
	
	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	
	// Contact info
	displayName := contact.GetDisplayName()
	if contact.Verified {
		displayName += " ✓"
	}
	
	lastSeen := "Last seen: " + formatLastSeen(contact.LastSeen)
	statusMessage := contact.StatusMessage
	if statusMessage == "" {
		statusMessage = "No status message"
	}
	
	// Format contact entry
	line1 := fmt.Sprintf("%s %s", statusStyle.Render(statusIcon), displayName)
	line2 := lipgloss.NewStyle().Foreground(c.theme.Secondary).Render(lastSeen)
	line3 := lipgloss.NewStyle().Foreground(c.theme.Secondary).Italic(true).Render(fmt.Sprintf("\"%s\"", statusMessage))
	
	content := lipgloss.JoinVertical(lipgloss.Left, line1, line2, line3)
	
	return style.Render(content)
}

// getVisibleContacts returns contacts that should be visible in the current scroll position
func (c *ContactsView) getVisibleContacts() []models.Contact {
	if len(c.contacts) == 0 {
		return []models.Contact{}
	}
	
	listHeight := c.height - 4
	maxContacts := listHeight / 4 // Each contact takes ~4 lines
	
	start := c.scrollOffset
	end := start + maxContacts
	
	if end > len(c.contacts) {
		end = len(c.contacts)
	}
	
	if start >= len(c.contacts) {
		start = len(c.contacts) - 1
	}
	
	return c.contacts[start:end]
}

// adjustScroll adjusts scroll position to keep selected item visible
func (c *ContactsView) adjustScroll() {
	listHeight := c.height - 4
	maxContacts := listHeight / 4
	
	if c.selectedIdx < c.scrollOffset {
		c.scrollOffset = c.selectedIdx
	} else if c.selectedIdx >= c.scrollOffset+maxContacts {
		c.scrollOffset = c.selectedIdx - maxContacts + 1
	}
	
	if c.scrollOffset < 0 {
		c.scrollOffset = 0
	}
}

// generateSampleContacts generates sample contacts for testing
func generateSampleContacts() []models.Contact {
	return []models.Contact{
		{
			UserID:        "alice_123",
			DisplayName:   "Alice Cooper",
			Status:        models.UserStatusOnline,
			StatusMessage: "Working on the auth system",
			Verified:      true,
		},
		{
			UserID:        "bob_456",
			DisplayName:   "Bob Wilson",
			Status:        models.UserStatusAway,
			StatusMessage: "Deploying to staging",
			Verified:      true,
		},
		{
			UserID:        "charlie_789",
			DisplayName:   "Charlie Davis",
			Status:        models.UserStatusOffline,
			StatusMessage: "In a meeting",
			Verified:      false,
		},
	}
}
