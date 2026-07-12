package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type MetaPane struct {
	Branch      string
	Upstream    string
	Ahead       int
	Behind      int
	LastFetch   time.Time
	StatusText  string
	Logs        []string
	MaxLogs     int
}

func NewMetaPane() *MetaPane {
	return &MetaPane{
		Branch:     "unknown",
		StatusText: "Initializing...",
		Logs:       []string{"System initialized"},
		MaxLogs:    10,
	}
}

func (mp *MetaPane) AddLog(msg string) {
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, msg)
	
	mp.Logs = append(mp.Logs, logEntry)
	if len(mp.Logs) > mp.MaxLogs {
		mp.Logs = mp.Logs[len(mp.Logs)-mp.MaxLogs:]
	}
}

func (mp *MetaPane) View(styles Styles, width int, height int) string {
	var sb strings.Builder

	// Style headers
	sectionHeader := func(text string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(ColorCyan)).Bold(true).Render(text)
	}

	// 1. Branch information
	sb.WriteString(sectionHeader("● REPO METRICS") + "\n")
	
	branchVal := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Bold(true).Render(mp.Branch)
	sb.WriteString(fmt.Sprintf("  Branch:   %s\n", branchVal))

	upstreamVal := "none"
	if mp.Upstream != "" {
		upstreamVal = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple)).Render(mp.Upstream)
	}
	sb.WriteString(fmt.Sprintf("  Upstream: %s\n", upstreamVal))

	// Ahead/Behind Divergence
	divStr := "Up to date"
	if mp.Ahead > 0 && mp.Behind > 0 {
		divStr = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange)).Render(fmt.Sprintf("▲ %d ahead, ▼ %d behind", mp.Ahead, mp.Behind))
	} else if mp.Ahead > 0 {
		divStr = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Render(fmt.Sprintf("▲ %d ahead", mp.Ahead))
	} else if mp.Behind > 0 {
		divStr = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed)).Render(fmt.Sprintf("▼ %d behind (pull needed)", mp.Behind))
	}
	sb.WriteString(fmt.Sprintf("  Status:   %s\n", divStr))

	// Last Fetch time
	fetchTimeStr := "never"
	if !mp.LastFetch.IsZero() {
		fetchTimeStr = mp.LastFetch.Format("15:04:05")
	}
	sb.WriteString(fmt.Sprintf("  Last check: %s\n", fetchTimeStr))
	
	// Activity status
	sb.WriteString(fmt.Sprintf("  Activity:   %s\n\n", lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow)).Render(mp.StatusText)))

	// 2. Activity log
	sb.WriteString(sectionHeader("● ACTIVITY FEED") + "\n")
	
	// Print logs starting from bottom (most recent) to fit height
	logSpace := height - 7 // Remaining lines for logs
	if logSpace < 1 {
		logSpace = 1
	}

	startIdx := len(mp.Logs) - logSpace
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(mp.Logs); i++ {
		logLine := mp.Logs[i]
		if len(logLine) > width {
			logLine = logLine[:width-3] + "..."
		}
		sb.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color(ColorComment)).Render(logLine) + "\n")
	}

	return sb.String()
}
