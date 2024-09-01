package menu

import "github.com/charmbracelet/lipgloss"

func newStyles() styles {
	return styles{
		errorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Bold(true),
	}
}

type styles struct {
	errorStyle lipgloss.Style
}
