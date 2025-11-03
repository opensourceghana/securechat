package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opensourceghana/securechat/internal/config"
	"github.com/opensourceghana/securechat/pkg/core"
	"github.com/opensourceghana/securechat/pkg/ui"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		showVersion = flag.Bool("version", false, "Show version information")
		debug      = flag.Bool("debug", false, "Enable debug mode")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("SecureChat %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if *debug {
		cfg.Debug = true
	}

	// Generate user ID if not set
	if cfg.User.ID == "" {
		cfg.User.ID = fmt.Sprintf("user_%d", time.Now().Unix())
	}

	// Initialize the core application
	coreApp, err := core.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize core application: %v", err)
	}
	defer coreApp.Close()

	// Initialize the TUI application
	uiApp := ui.NewApp(cfg)
	
	// Connect core app to UI (simplified integration)
	// In a full implementation, we'd have proper event channels
	
	// Try to connect to network
	if len(cfg.Network.RelayServers) > 0 {
		log.Printf("Connecting to relay server...")
		if err := coreApp.Connect(); err != nil {
			log.Printf("Warning: Failed to connect to relay server: %v", err)
		}
	}

	// Create Bubble Tea program
	p := tea.NewProgram(
		uiApp,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}

func loadConfig(configPath string) (*config.Config, error) {
	if configPath == "" {
		// Try default locations
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		candidates := []string{
			filepath.Join(homeDir, ".config", "securechat", "config.yaml"),
			filepath.Join(homeDir, ".securechat.yaml"),
			"config.yaml",
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				configPath = candidate
				break
			}
		}
	}

	if configPath == "" {
		// Use default configuration
		return config.Default(), nil
	}

	return config.LoadFromFile(configPath)
}
