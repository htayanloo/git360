package tui

import "github.com/charmbracelet/lipgloss"

// Theme colors (Dracula-inspired palette for high contrast and modern styling)
const (
	ColorBackground = "#282a36"
	ColorCurrent    = "#44475a"
	ColorForeground = "#f8f8f2"
	ColorComment    = "#6272a4"
	ColorCyan       = "#8be9fd"
	ColorGreen      = "#50fa7b"
	ColorOrange     = "#ffb86c"
	ColorPink       = "#ff79c6"
	ColorPurple     = "#bd93f9"
	ColorRed        = "#ff5555"
	ColorYellow     = "#f1fa8c"
	ColorMutedGray  = "#444444"
)

// Pane identifies which split screen panel currently has keyboard focus.
type Pane int

const (
	PaneGraph Pane = iota
	PaneStatus
	PaneDiff
)

type Styles struct {
	// Base application layout styles
	HeaderStyle    lipgloss.Style
	FooterStyle    lipgloss.Style
	HelpKeyStyle   lipgloss.Style
	HelpDescStyle  lipgloss.Style
	
	// Pane container styles
	ActivePaneBorder   lipgloss.Style
	InactivePaneBorder lipgloss.Style
	PaneTitleActive    lipgloss.Style
	PaneTitleInactive  lipgloss.Style
	
	// List and navigation styling
	SelectedLineStyle  lipgloss.Style
	NormalLineStyle    lipgloss.Style
	
	// Git status visual mappings
	StagedFileStyle     lipgloss.Style
	UnstagedFileStyle   lipgloss.Style
	UntrackedFileStyle  lipgloss.Style
	ConflictedFileStyle lipgloss.Style
	
	// Diff highlighter styling
	DiffHeaderStyle lipgloss.Style
	DiffAddedStyle  lipgloss.Style
	DiffRemovedStyle lipgloss.Style
	
	// Graph line coloring (cyclic)
	GraphColors []lipgloss.Style
}

func DefaultStyles() Styles {
	s := Styles{}

	// Header and Footer top/bottom bars
	s.HeaderStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorCurrent)).
		Foreground(lipgloss.Color(ColorCyan)).
		Bold(true).
		Padding(0, 1)

	s.FooterStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorCurrent)).
		Foreground(lipgloss.Color(ColorForeground)).
		Height(1)

	s.HelpKeyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPink)).
		Bold(true)

	s.HelpDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorComment))

	// Borders for active vs. inactive panels
	s.ActivePaneBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPurple)).
		Padding(0, 1)

	s.InactivePaneBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorMutedGray)).
		Padding(0, 1)

	s.PaneTitleActive = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPurple)).
		Bold(true)

	s.PaneTitleInactive = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorComment))

	// Navigation lines
	s.SelectedLineStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorCurrent)).
		Foreground(lipgloss.Color(ColorForeground)).
		Bold(true)

	s.NormalLineStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorForeground))

	// Git files list
	s.StagedFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen))
	s.UnstagedFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange))
	s.UntrackedFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorComment))
	s.ConflictedFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed)).Bold(true)

	// Diff Viewer syntax styling
	s.DiffHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorCyan))
	s.DiffAddedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen))
	s.DiffRemovedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed))

	// Dynamic colors for Git graph lines
	s.GraphColors = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorCyan)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPink)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed)),
	}

	return s
}
