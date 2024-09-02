package menu

import (
	"io/fs"
	"net/url"
	"os"
	"path"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/gabe565/cli-of-life/internal/game/util"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/gabe565/cli-of-life/internal/pattern/embedded"
)

const (
	sourceEmbedded = "embedded"
	sourceFile     = "file"
	sourceURL      = "url"
)

func (m *Menu) initForm() tea.Cmd {
	m.form = m.form.WithWidth(lipgloss.Width(Title))
	cmds := []tea.Cmd{m.form.Init()}

	form, cmd := m.form.Update(m.size)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

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
				AllowedTypes(pattern.Extensions()).
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

		options = append(options, huh.NewOption(p.NameAuthor(), "embedded://"+path))
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

func (m *Menu) choosePatternForm(urls []string) tea.Cmd {
	opts := make([]huh.Option[string], 0, len(urls))
	for _, u := range urls {
		opts = append(opts, huh.NewOption(path.Base(u), u))
	}

	m.form = util.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose Pattern").
				Options(opts...).
				Value(&m.config.Pattern),
		),
	)
	return m.initForm()
}
