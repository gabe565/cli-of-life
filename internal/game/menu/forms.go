package menu

import (
	"net/url"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/gabe565/cli-of-life/internal/game/util"
	"github.com/gabe565/cli-of-life/internal/pattern"
)

const (
	sourceFile = "file"
	sourceURL  = "url"
)

func (m *Menu) loadPatternForm() tea.Cmd {
	m.form = util.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pattern Source").
				Options(
					huh.NewOption("File Picker", sourceFile),
					huh.NewOption("Path/URL", sourceURL),
				).
				Value(&m.patternSrc),
		),
	)
	return m.initForm()
}

func (m *Menu) patternFileForm() tea.Cmd {
	wd, _ := os.Getwd()

	m.form = util.NewForm(
		huh.NewGroup(
			huh.NewFilePicker().
				Title("Pattern File Picker").
				Picking(true).
				CurrentDirectory(wd).
				ShowSize(true).
				ShowPermissions(false).
				AllowedTypes([]string{pattern.ExtRLE, pattern.ExtPlaintext}).
				Height(15).
				Value(&m.config.Pattern),
		),
	)
	return m.initForm()
}

func (m *Menu) patternURLForm() tea.Cmd {
	ne := huh.ValidateNotEmpty()
	m.form = util.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Pattern Path/URL").
				Validate(func(s string) error {
					if err := ne(s); err != nil {
						return err
					}
					_, err := url.Parse(s)
					return err
				}).
				Value(&m.config.Pattern),
		),
	)
	return m.initForm()
}
