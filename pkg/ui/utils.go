package ui

import (
	"fmt"
	"time"
)

// formatLastSeen formats a timestamp into a human-readable "last seen" string
func formatLastSeen(lastSeen time.Time) string {
	now := time.Now()
	duration := now.Sub(lastSeen)
	
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		return lastSeen.Format("Jan 2, 2006")
	}
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	
	if maxLen <= 3 {
		return s[:maxLen]
	}
	
	return s[:maxLen-3] + "..."
}

// centerString centers a string within the given width
func centerString(s string, width int) string {
	if len(s) >= width {
		return s
	}
	
	padding := width - len(s)
	leftPad := padding / 2
	rightPad := padding - leftPad
	
	return fmt.Sprintf("%*s%s%*s", leftPad, "", s, rightPad, "")
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	
	words := splitWords(text)
	if len(words) == 0 {
		return []string{""}
	}
	
	var lines []string
	var currentLine string
	
	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	return lines
}

// splitWords splits text into words, preserving whitespace
func splitWords(text string) []string {
	var words []string
	var currentWord string
	
	for _, char := range text {
		if char == ' ' || char == '\t' || char == '\n' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(char)
		}
	}
	
	if currentWord != "" {
		words = append(words, currentWord)
	}
	
	return words
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// formatDuration formats a duration to a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

// isValidUserID checks if a user ID is valid
func isValidUserID(userID string) bool {
	if len(userID) < 3 || len(userID) > 50 {
		return false
	}
	
	// Check for valid characters (alphanumeric, underscore, hyphen)
	for _, char := range userID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}
	
	return true
}

// sanitizeInput sanitizes user input by removing control characters
func sanitizeInput(input string) string {
	var result []rune
	
	for _, char := range input {
		// Allow printable characters and common whitespace
		if char >= 32 && char <= 126 || char == '\t' || char == '\n' {
			result = append(result, char)
		}
	}
	
	return string(result)
}

// generateProgressBar generates a text-based progress bar
func generateProgressBar(current, total int, width int) string {
	if total == 0 || width <= 0 {
		return ""
	}
	
	percentage := float64(current) / float64(total)
	if percentage > 1.0 {
		percentage = 1.0
	}
	
	filled := int(percentage * float64(width))
	empty := width - filled
	
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	
	return bar
}
