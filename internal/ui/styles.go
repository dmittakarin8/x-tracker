package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
// Base colors
subtle    = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#4A4A4A"}
highlight = lipgloss.AdaptiveColor{Light: "#7B61FF", Dark: "#9D86FF"}
special   = lipgloss.AdaptiveColor{Light: "#00CC6A", Dark: "#00FF84"}

// Styles
titleStyle = lipgloss.NewStyle().
    Foreground(highlight).
    Bold(true).
    Padding(0, 0).
    MarginBottom(0)

statusBarStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#E2E2E2")).
    Background(lipgloss.Color("#1A1B26")).
    MarginTop(1).
    Padding(0, 0)

errorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FF5555")).
    Bold(true)

listStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(subtle).
    BorderStyle(lipgloss.RoundedBorder()).
    Padding(1).
    MarginTop(1)

itemStyle = lipgloss.NewStyle().
    PaddingLeft(4).
    Foreground(lipgloss.Color("#CCCCCC"))

selectedItemStyle = lipgloss.NewStyle().
    PaddingLeft(2).
    Foreground(special).
    Bold(true).
    SetString("→ ")

// Input field styles
inputPromptStyle = lipgloss.NewStyle().
    Foreground(highlight).
    Bold(true).
    PaddingRight(1)

inputStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FFFFFF")).
    Background(lipgloss.Color("#2D2D3A")).
    Padding(0, 1)

placeholderStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#808080")).
    Italic(true)

cursorStyle = lipgloss.NewStyle().
    Foreground(highlight).
    Bold(true)

focusedInputStyle = inputStyle.Copy().
    BorderStyle(lipgloss.NormalBorder()).
    BorderForeground(highlight).
    Background(lipgloss.Color("#363646"))

helpStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#ABABAB")).
    MarginTop(1)

removePromptStyle = lipgloss.NewStyle().
    Foreground(highlight).
    Bold(true)
) 