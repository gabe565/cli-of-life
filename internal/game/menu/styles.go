package menu

import "charm.land/lipgloss/v2"

func newStyles() styles {
	return styles{
		errorStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("204")).
			Bold(true),
	}
}

type styles struct {
	errorStyle lipgloss.Style
}
