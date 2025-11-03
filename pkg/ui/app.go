package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opensourceghana/securechat/internal/config"
)

// App represents the main TUI application
type App struct {
	config *config.Config
	width  int
	height int
	
	// Current view state
	currentView ViewType
	views       map[ViewType]tea.Model
	
	// Global state
	theme *Theme
}

// ViewType represents different views in the application
type ViewType string

const (
	ViewChat     ViewType = "chat"
	ViewContacts ViewType = "contacts"
	ViewSettings ViewType = "settings"
	ViewHelp     ViewType = "help"
)

// Theme contains styling information
type Theme struct {
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Background  lipgloss.Color
	Foreground  lipgloss.Color
	Border      lipgloss.Color
	Highlight   lipgloss.Color
	Error       lipgloss.Color
	Warning     lipgloss.Color
	Success     lipgloss.Color
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config) *App {
	app := &App{
		config:      cfg,
		currentView: ViewChat,
		views:       make(map[ViewType]tea.Model),
		theme:       getTheme(cfg.UI.Theme),
	}
	
	// Initialize views
	app.views[ViewChat] = NewChatView(cfg, app.theme)
	app.views[ViewContacts] = NewContactsView(cfg, app.theme)
	app.views[ViewSettings] = NewSettingsView(cfg, app.theme)
	app.views[ViewHelp] = NewHelpView(cfg, app.theme)
	
	return app
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		a.views[a.currentView].Init(),
	)
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		
		// Update all views with new size
		for viewType, view := range a.views {
			a.views[viewType], _ = view.Update(msg)
		}
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
			return a, tea.Quit
			
		case "ctrl+c":
			return a, tea.Quit
			
		case "ctrl+,":
			a.currentView = ViewSettings
			return a, a.views[a.currentView].Init()
			
		case "ctrl+/":
			a.currentView = ViewHelp
			return a, a.views[a.currentView].Init()
			
		case "f1":
			a.currentView = ViewChat
			return a, a.views[a.currentView].Init()
			
		case "f2":
			a.currentView = ViewContacts
			return a, a.views[a.currentView].Init()
			
		case "esc":
			// Return to chat view from other views
			if a.currentView != ViewChat {
				a.currentView = ViewChat
				return a, a.views[a.currentView].Init()
			}
		}
	}
	
	// Update current view
	a.views[a.currentView], cmd = a.views[a.currentView].Update(msg)
	
	return a, cmd
}

// View implements tea.Model
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}
	
	// Render current view
	content := a.views[a.currentView].View()
	
	// Add status bar
	statusBar := a.renderStatusBar()
	
	// Combine content and status bar
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		statusBar,
	)
}

// renderStatusBar renders the bottom status bar
func (a *App) renderStatusBar() string {
	style := lipgloss.NewStyle().
		Background(a.theme.Primary).
		Foreground(a.theme.Background).
		Padding(0, 1)
	
	// Status indicators
	status := "● Online"
	if a.currentView != ViewChat {
		status += " | Press Esc to return to chat"
	}
	
	// Keyboard shortcuts
	shortcuts := "F1: Chat | F2: Contacts | Ctrl+,: Settings | Ctrl+/: Help | Ctrl+Q: Quit"
	
	// Create status bar with proper width
	leftPart := style.Render(status)
	rightPart := style.Render(shortcuts)
	
	padding := a.width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart)
	if padding < 0 {
		padding = 0
	}
	
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftPart,
		style.Render(lipgloss.PlaceHorizontal(padding, lipgloss.Right, "")),
		rightPart,
	)
}

// getTheme returns the appropriate theme based on the theme name
func getTheme(themeName string) *Theme {
	switch themeName {
	case "light":
		return &Theme{
			Primary:    lipgloss.Color("#0066cc"),
			Secondary:  lipgloss.Color("#6c757d"),
			Background: lipgloss.Color("#ffffff"),
			Foreground: lipgloss.Color("#2d2d2d"),
			Border:     lipgloss.Color("#d0d0d0"),
			Highlight:  lipgloss.Color("#f5f5f5"),
			Error:      lipgloss.Color("#dc3545"),
			Warning:    lipgloss.Color("#ffc107"),
			Success:    lipgloss.Color("#28a745"),
		}
	default: // "dark" or "auto"
		return &Theme{
			Primary:    lipgloss.Color("#00d4aa"),
			Secondary:  lipgloss.Color("#6c757d"),
			Background: lipgloss.Color("#1a1a1a"),
			Foreground: lipgloss.Color("#e0e0e0"),
			Border:     lipgloss.Color("#404040"),
			Highlight:  lipgloss.Color("#2d2d2d"),
			Error:      lipgloss.Color("#ff6b6b"),
			Warning:    lipgloss.Color("#ffd93d"),
			Success:    lipgloss.Color("#6bcf7f"),
		}
	}
}
