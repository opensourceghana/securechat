package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opensourceghana/securechat/internal/config"
)

// SettingsView represents the settings interface
type SettingsView struct {
	config   *config.Config
	theme    *Theme
	width    int
	height   int
	
	// Settings state
	selectedSection int
	selectedItem    int
	editMode        bool
	editValue       string
	
	// Settings sections
	sections []SettingsSection
}

// SettingsSection represents a group of related settings
type SettingsSection struct {
	Name  string
	Items []SettingsItem
}

// SettingsItem represents a single setting
type SettingsItem struct {
	Name        string
	Value       interface{}
	Type        SettingsType
	Options     []string
	Description string
	ReadOnly    bool
}

// SettingsType represents the type of a setting
type SettingsType int

const (
	SettingsTypeString SettingsType = iota
	SettingsTypeBool
	SettingsTypeSelect
	SettingsTypeInt
	SettingsTypeButton
)

// NewSettingsView creates a new settings view
func NewSettingsView(cfg *config.Config, theme *Theme) *SettingsView {
	view := &SettingsView{
		config: cfg,
		theme:  theme,
	}
	
	view.sections = view.buildSettingsSections()
	
	return view
}

// Init implements tea.Model
func (s *SettingsView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (s *SettingsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height - 2 // Account for status bar
		
	case tea.KeyMsg:
		if s.editMode {
			return s.handleEditInput(msg)
		}
		
		switch msg.String() {
		case "up", "k":
			s.navigateUp()
			
		case "down", "j":
			s.navigateDown()
			
		case "left", "h":
			if s.selectedSection > 0 {
				s.selectedSection--
				s.selectedItem = 0
			}
			
		case "right", "l":
			if s.selectedSection < len(s.sections)-1 {
				s.selectedSection++
				s.selectedItem = 0
			}
			
		case "tab":
			s.selectedSection = (s.selectedSection + 1) % len(s.sections)
			s.selectedItem = 0
			
		case "enter":
			s.activateCurrentItem()
			
		case "space":
			s.toggleCurrentItem()
		}
	}
	
	return s, nil
}

// View implements tea.Model
func (s *SettingsView) View() string {
	if s.width == 0 || s.height == 0 {
		return "Loading settings..."
	}
	
	// Create main layout
	header := s.renderHeader()
	content := s.renderContent()
	footer := s.renderFooter()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

// handleEditInput handles keyboard input during edit mode
func (s *SettingsView) handleEditInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		s.editMode = false
		s.editValue = ""
		
	case "enter":
		s.saveEditValue()
		s.editMode = false
		s.editValue = ""
		
	case "backspace":
		if len(s.editValue) > 0 {
			s.editValue = s.editValue[:len(s.editValue)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			s.editValue += msg.String()
		}
	}
	
	return s, nil
}

// renderHeader renders the settings view header
func (s *SettingsView) renderHeader() string {
	style := lipgloss.NewStyle().
		Background(s.theme.Primary).
		Foreground(s.theme.Background).
		Padding(0, 1).
		Width(s.width)
	
	title := "Settings"
	
	return style.Render(title)
}

// renderContent renders the main settings content
func (s *SettingsView) renderContent() string {
	contentHeight := s.height - 3 // Account for header and footer
	
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(s.theme.Border).
		Width(s.width).
		Height(contentHeight).
		Padding(1)
	
	var sections []string
	
	for i, section := range s.sections {
		sectionContent := s.renderSection(section, i == s.selectedSection)
		sections = append(sections, sectionContent)
	}
	
	content := strings.Join(sections, "\n\n")
	
	return style.Render(content)
}

// renderSection renders a single settings section
func (s *SettingsView) renderSection(section SettingsSection, isSelected bool) string {
	var titleStyle lipgloss.Style
	if isSelected {
		titleStyle = lipgloss.NewStyle().
			Foreground(s.theme.Primary).
			Bold(true)
	} else {
		titleStyle = lipgloss.NewStyle().
			Foreground(s.theme.Secondary).
			Bold(true)
	}
	
	title := titleStyle.Render(section.Name)
	
	var items []string
	for i, item := range section.Items {
		itemSelected := isSelected && i == s.selectedItem
		itemContent := s.renderItem(item, itemSelected)
		items = append(items, itemContent)
	}
	
	itemsContent := strings.Join(items, "\n")
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		itemsContent,
	)
}

// renderItem renders a single settings item
func (s *SettingsView) renderItem(item SettingsItem, isSelected bool) string {
	var style lipgloss.Style
	if isSelected {
		style = lipgloss.NewStyle().
			Background(s.theme.Highlight).
			Foreground(s.theme.Foreground).
			Padding(0, 1)
	} else {
		style = lipgloss.NewStyle().
			Foreground(s.theme.Foreground).
			Padding(0, 1)
	}
	
	var valueStr string
	switch item.Type {
	case SettingsTypeBool:
		if item.Value.(bool) {
			valueStr = "[●] Yes  [ ] No"
		} else {
			valueStr = "[ ] Yes  [●] No"
		}
		
	case SettingsTypeSelect:
		current := item.Value.(string)
		var options []string
		for _, opt := range item.Options {
			if opt == current {
				options = append(options, "[●] "+opt)
			} else {
				options = append(options, "[ ] "+opt)
			}
		}
		valueStr = strings.Join(options, "  ")
		
	case SettingsTypeString:
		if s.editMode && isSelected {
			valueStr = fmt.Sprintf("[%s│]", s.editValue)
		} else {
			valueStr = fmt.Sprintf("[%s]", item.Value.(string))
		}
		
	case SettingsTypeInt:
		valueStr = fmt.Sprintf("[%d]", item.Value.(int))
		
	case SettingsTypeButton:
		valueStr = "[" + item.Value.(string) + "]"
		
	default:
		valueStr = fmt.Sprintf("%v", item.Value)
	}
	
	if item.ReadOnly {
		valueStr += " (read-only)"
	}
	
	content := fmt.Sprintf("├─ %s: %s", item.Name, valueStr)
	
	return style.Render(content)
}

// renderFooter renders the settings view footer
func (s *SettingsView) renderFooter() string {
	style := lipgloss.NewStyle().
		Background(s.theme.Secondary).
		Foreground(s.theme.Background).
		Padding(0, 1).
		Width(s.width)
	
	shortcuts := "[Tab] Next section  [Enter] Edit  [Space] Toggle  [Esc] Back to chat"
	
	return style.Render(shortcuts)
}

// buildSettingsSections creates the settings sections from config
func (s *SettingsView) buildSettingsSections() []SettingsSection {
	return []SettingsSection{
		{
			Name: "Profile",
			Items: []SettingsItem{
				{
					Name:  "Display Name",
					Value: s.config.User.DisplayName,
					Type:  SettingsTypeString,
				},
				{
					Name:  "Status Message",
					Value: s.config.User.StatusMessage,
					Type:  SettingsTypeString,
				},
				{
					Name:     "User ID",
					Value:    s.config.User.ID,
					Type:     SettingsTypeString,
					ReadOnly: true,
				},
			},
		},
		{
			Name: "Security",
			Items: []SettingsItem{
				{
					Name:    "Auto-accept keys",
					Value:   s.config.Security.AutoAcceptKeys,
					Type:    SettingsTypeBool,
					Options: []string{"No", "Ask", "Yes"},
				},
				{
					Name:  "Message retention",
					Value: fmt.Sprintf("%d days", s.config.Security.MessageRetentionDays),
					Type:  SettingsTypeSelect,
					Options: []string{"1 day", "7 days", "30 days", "90 days", "1 year", "Forever"},
				},
				{
					Name:  "Export keys",
					Value: "Export...",
					Type:  SettingsTypeButton,
				},
				{
					Name:  "Import keys",
					Value: "Import...",
					Type:  SettingsTypeButton,
				},
			},
		},
		{
			Name: "Interface",
			Items: []SettingsItem{
				{
					Name:    "Theme",
					Value:   s.config.UI.Theme,
					Type:    SettingsTypeSelect,
					Options: []string{"dark", "light", "auto"},
				},
				{
					Name:  "Notifications",
					Value: s.config.UI.Notifications,
					Type:  SettingsTypeBool,
				},
				{
					Name:  "Sound alerts",
					Value: s.config.UI.SoundEnabled,
					Type:  SettingsTypeBool,
				},
				{
					Name:    "Timestamp format",
					Value:   s.config.UI.TimestampFormat,
					Type:    SettingsTypeSelect,
					Options: []string{"15:04", "3:04 PM", "15:04:05"},
				},
			},
		},
		{
			Name: "Network",
			Items: []SettingsItem{
				{
					Name:  "Relay servers",
					Value: strings.Join(s.config.Network.RelayServers, ", "),
					Type:  SettingsTypeString,
				},
				{
					Name:  "P2P connections",
					Value: s.config.Network.P2PEnabled,
					Type:  SettingsTypeBool,
				},
				{
					Name:    "Connection timeout",
					Value:   s.config.Network.ConnectionTimeout.String(),
					Type:    SettingsTypeSelect,
					Options: []string{"10s", "30s", "60s", "120s"},
				},
			},
		},
	}
}

// navigateUp moves selection up within current section
func (s *SettingsView) navigateUp() {
	if s.selectedItem > 0 {
		s.selectedItem--
	} else if s.selectedSection > 0 {
		s.selectedSection--
		s.selectedItem = len(s.sections[s.selectedSection].Items) - 1
	}
}

// navigateDown moves selection down within current section
func (s *SettingsView) navigateDown() {
	if s.selectedItem < len(s.sections[s.selectedSection].Items)-1 {
		s.selectedItem++
	} else if s.selectedSection < len(s.sections)-1 {
		s.selectedSection++
		s.selectedItem = 0
	}
}

// activateCurrentItem activates the currently selected item
func (s *SettingsView) activateCurrentItem() {
	item := &s.sections[s.selectedSection].Items[s.selectedItem]
	
	if item.ReadOnly {
		return
	}
	
	switch item.Type {
	case SettingsTypeString:
		s.editMode = true
		s.editValue = item.Value.(string)
		
	case SettingsTypeButton:
		// TODO: Handle button actions
		
	default:
		s.toggleCurrentItem()
	}
}

// toggleCurrentItem toggles the value of the current item
func (s *SettingsView) toggleCurrentItem() {
	item := &s.sections[s.selectedSection].Items[s.selectedItem]
	
	if item.ReadOnly {
		return
	}
	
	switch item.Type {
	case SettingsTypeBool:
		item.Value = !item.Value.(bool)
		s.updateConfig(item)
		
	case SettingsTypeSelect:
		// Cycle through options
		current := item.Value.(string)
		for i, opt := range item.Options {
			if opt == current {
				nextIdx := (i + 1) % len(item.Options)
				item.Value = item.Options[nextIdx]
				s.updateConfig(item)
				break
			}
		}
	}
}

// saveEditValue saves the edited value
func (s *SettingsView) saveEditValue() {
	item := &s.sections[s.selectedSection].Items[s.selectedItem]
	item.Value = s.editValue
	s.updateConfig(item)
}

// updateConfig updates the configuration based on the changed item
func (s *SettingsView) updateConfig(item *SettingsItem) {
	// TODO: Update the actual config and save to file
	// This is a simplified version - in a real implementation,
	// we'd need to map settings back to config fields
}
