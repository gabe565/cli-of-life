package menu

import (
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path"
	"strings"

	"gabe565.com/cli-of-life/internal/game/util"
	"gabe565.com/cli-of-life/internal/pattern"
	"gabe565.com/cli-of-life/internal/pattern/embedded"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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

		p, err := pattern.UnmarshalRLE(f)
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

var ErrLineBreak = errors.New("line breaks are not allowed")

type trimSpaceAccessor struct {
	value *string
}

func (t *trimSpaceAccessor) Get() string {
	return *t.value
}

func (t *trimSpaceAccessor) Set(value string) {
	*t.value = strings.TrimSpace(value)
}

func (m *Menu) patternURLForm() tea.Cmd {
	ne := huh.ValidateNotEmpty()
	m.form = util.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Pattern Path/URL").
				Description("  Supports direct rle/cells URLs\n  or web pages with rle/cells links.\n").
				Validate(func(s string) error {
					if err := ne(s); err != nil {
						return err
					}

					s = strings.TrimSpace(s)
					if strings.Contains(s, "\n") {
						return ErrLineBreak
					}

					_, err := url.Parse(s)
					var urlErr *url.Error
					if errors.As(err, &urlErr) {
						return urlErr.Err
					}
					return err
				}).
				Accessor(&trimSpaceAccessor{&m.config.Pattern}),
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
