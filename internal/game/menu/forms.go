package menu

import (
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/gabe565/cli-of-life/internal/game/util"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/gabe565/cli-of-life/internal/pattern/embedded"
)

const (
	sourceEmbedded = "embedded"
	sourceFile     = "file"
	sourceURL      = "url"
)

func (m *Menu) loadPatternForm() tea.Cmd {
	m.form = util.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pattern Source").
				Options(
					huh.NewOption("Embedded Patterns", sourceEmbedded),
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

func (m *Menu) patternEmbeddedForm() tea.Cmd {
	var options []huh.Option[string]
	if err := fs.WalkDir(embedded.Embedded, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		f, err := embedded.Embedded.Open(path)
		if err != nil {
			return err
		}

		p, err := pattern.Unmarshal(f)
		if err != nil {
			return err
		}

		var name string
		if p.Name != "" {
			name = p.Name
			if p.Author != "" {
				name += " by " + p.Author
			}
		} else {
			name = filepath.Base(path)
		}

		options = append(options, huh.NewOption(name, "embedded://"+path))
		return nil
	}); err != nil {
		m.error = err
		return nil
	}

	m.form = util.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Embedded Pattern Picker").
				Options(options...).
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
