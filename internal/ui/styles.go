package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors - Tokyo Night palette (refined)
	ColorBgDark    = lipgloss.Color("#1a1b26")
	ColorBgSidebar = lipgloss.Color("#24283b")
	ColorBgInput   = lipgloss.Color("#414868")
	ColorBgActive  = lipgloss.Color("#292e42")
	ColorBorder    = lipgloss.Color("#3b4261")
	ColorBorderDim = lipgloss.Color("#2f3549")
	ColorText      = lipgloss.Color("#c0caf5")
	ColorTextMuted = lipgloss.Color("#565f89")
	ColorTextDim   = lipgloss.Color("#444b6a")
	ColorBlue      = lipgloss.Color("#7aa2f7")
	ColorCyan      = lipgloss.Color("#7dcfff")
	ColorGreen     = lipgloss.Color("#9ece6a")
	ColorYellow    = lipgloss.Color("#e0af68")
	ColorRed       = lipgloss.Color("#f7768e")
	ColorPurple    = lipgloss.Color("#bb9af7")
	ColorOrange    = lipgloss.Color("#ff9e64")

	// Card / panel border style
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	CardActiveStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBlue)

	CardDimStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderDim)

	// Header styles
	HeaderTitleStyle = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Bold(true).
				Padding(0, 1)

	HeaderTaglineStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted).
				Italic(true)

	HeaderTabStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Padding(0, 1)

	HeaderTabActiveStyle = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Bold(true).
				Underline(true).
				Padding(0, 1)

	HeaderHelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Padding(0, 1)

	AdminWarningStyle = lipgloss.NewStyle().
				Foreground(ColorRed).
				Bold(true)

	// Sidebar styles
	SidebarTitleStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted).
				Bold(true)

	SidebarItemStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted).
				Padding(0, 1)

	SidebarItemActiveStyle = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Bold(true).
				Padding(0, 1)

	SidebarItemActiveBgStyle = lipgloss.NewStyle().
					Foreground(ColorBlue).
					Background(ColorBgActive).
					Bold(true).
					Padding(0, 1)

	SidebarItemCountStyle = lipgloss.NewStyle().
				Foreground(ColorTextDim)

	// Command list styles
	CmdHeaderStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Bold(true)

	CmdSyntaxStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	CmdDescStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	CmdMatchStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true)

	CmdCursorStyle = lipgloss.NewStyle().
			Foreground(ColorBlue).
			Bold(true)

	// Input styles
	InputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	InputBoxFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBlue)

	InputPromptStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Bold(true)

	// Preview styles
	PreviewHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Bold(true)

	PreviewSectionStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted).
				Bold(true)

	PreviewCmdStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	PreviewFlagStyle = lipgloss.NewStyle().
				Foreground(ColorCyan)

	PreviewDescStyle = lipgloss.NewStyle().
				Foreground(ColorText)

	PreviewExampleStyle = lipgloss.NewStyle().
				Foreground(ColorYellow)

	PreviewDefaultStyle = lipgloss.NewStyle().
				Foreground(ColorTextDim)

	// Placeholder editing styles
	PlaceholderEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorYellow)

	PlaceholderFilledStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Bold(true)

	PlaceholderActiveStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Background(ColorBgActive).
				Bold(true)

	PlaceholderLabelStyle = lipgloss.NewStyle().
				Foreground(ColorCyan)

	PickerHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorPurple).
				Bold(true)

	PickerSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorGreen)

	PlaceholderWarnStyle = lipgloss.NewStyle().
				Foreground(ColorRed).
				Bold(true)

	// Badge styles
	BadgeBeginnerStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Bold(true)

	BadgeIntermediateStyle = lipgloss.NewStyle().
				Foreground(ColorYellow).
				Bold(true)

	BadgeAdvancedStyle = lipgloss.NewStyle().
				Foreground(ColorRed).
				Bold(true)

	// Output styles
	OutputHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Bold(true)

	OutputJsonKeyStyle = lipgloss.NewStyle().
				Foreground(ColorPurple)

	OutputJsonStringStyle = lipgloss.NewStyle().
				Foreground(ColorGreen)

	OutputJsonNumberStyle = lipgloss.NewStyle().
				Foreground(ColorYellow)

	OutputErrorStyle = lipgloss.NewStyle().
				Foreground(ColorRed).
				Bold(true)

	OutputSuccessStyle = lipgloss.NewStyle().
				Foreground(ColorGreen)

	// Learn styles
	LearnTopicStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1)

	LearnTopicActiveStyle = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Bold(true).
				Padding(0, 1)

	// Status bar / hints
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	StatusBarKeyStyle = lipgloss.NewStyle().
				Foreground(ColorCyan).
				Bold(true)

	// Tooltip styles
	TooltipStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPurple).
			Background(ColorBgSidebar).
			Padding(1, 2)

	// Scroll indicator
	ScrollStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim)

	// Section label (for "COMMANDS", "FLAGS", etc.)
	SectionLabelStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted).
				Bold(true)

	// Action hint strip (bottom of cards)
	ActionHintStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	ActionHintKeyStyle = lipgloss.NewStyle().
				Foreground(ColorCyan).
				Bold(true)
)

func GetDifficultyStyle(difficulty string) lipgloss.Style {
	switch difficulty {
	case "beginner":
		return BadgeBeginnerStyle
	case "intermediate":
		return BadgeIntermediateStyle
	case "advanced":
		return BadgeAdvancedStyle
	default:
		return lipgloss.NewStyle()
	}
}

func GetDifficultyBadge(difficulty string) string {
	style := GetDifficultyStyle(difficulty)
	if difficulty == "beginner" || difficulty == "intermediate" || difficulty == "advanced" {
		return style.Render("●")
	}
	return ""
}

func GetDifficultyLabel(difficulty string) string {
	style := GetDifficultyStyle(difficulty)
	switch difficulty {
	case "beginner":
		return style.Render("Beginner")
	case "intermediate":
		return style.Render("Intermediate")
	case "advanced":
		return style.Render("Advanced")
	default:
		return ""
	}
}

func GetCategoryIcon(category string) string {
	switch category {
	case "Container":
		return "□"
	case "Image":
		return "◇"
	case "Network":
		return "◎"
	case "Volume":
		return "◐"
	case "Session":
		return "☰"
	case "System":
		return "⊙"
	case "Registry":
		return "◈"
	default:
		return "·"
	}
}
