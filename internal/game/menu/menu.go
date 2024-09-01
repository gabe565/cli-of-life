package menu

import (
	"net/url"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/game/commands"
	"github.com/gabe565/cli-of-life/internal/game/components/buttons"
	"github.com/gabe565/cli-of-life/internal/game/conway"
	"github.com/gabe565/cli-of-life/internal/game/util"
	"github.com/gabe565/cli-of-life/internal/pattern"
)

const (
	Title = `
        ██  ██                ██      ██  ██    ██        
        ██                  ██  ██    ██      ██  ██  ██  
  ████  ██  ██       ███    ██        ██  ██  ██    ██  ██
██      ██  ██     ██   ██  ████      ██  ██  ████  ██████
██      ██  ██     ██   ██  ██        ██  ██  ██    ██    
  ████  ██  ██       ███    ██        ██  ██  ██      ███ 

`

	BtnResume = "Resume Game"
	BtnReset  = "Reset Game"
	BtnNew    = "New Game"
	BtnLoad   = "Load Pattern"
	BtnQuit   = "Quit"
)

func NewMenu(conf *config.Config, conway *conway.Conway) *Menu {
	m := &Menu{
		config: conf,
		keymap: newKeymap(),
		help:   help.New(),
		styles: newStyles(),

		conway:  conway,
		buttons: buttons.New(BtnResume, BtnReset, BtnNew, BtnLoad, BtnQuit),
	}
	m.buttons.List[0].Hidden = true
	m.buttons.List[1].Hidden = true
	return m
}

type Menu struct {
	config *config.Config
	size   tea.WindowSizeMsg
	keymap keymap
	help   help.Model
	styles styles

	conway     *conway.Conway
	buttons    *buttons.Buttons
	form       *huh.Form
	patternSrc string

	error error
}

func (m *Menu) Init() tea.Cmd { return nil }

const (
	sourceFile = "file"
	sourceURL  = "url"
)

func (m *Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = msg
	case tea.KeyMsg:
		switch { //nolint:gocritic
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		}
	}

	if m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}

		switch m.form.State {
		case huh.StateCompleted:
			m.form = nil

			switch m.patternSrc {
			case sourceFile:
				m.patternSrc = ""

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
				return m, m.initForm()
			case sourceURL:
				m.patternSrc = ""
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
				return m, m.initForm()
			default:
				if p, err := pattern.New(m.config); err == nil {
					m.conway.Pattern = p
					m.conway.ResetView()
					return m, commands.ChangeView(commands.Conway)
				} else {
					m.error = err
				}
			}
		case huh.StateAborted:
			m.form = nil
			return m, nil
		default:
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case commands.View:
		filled := m.conway.Pattern.Tree.FilledCoords()
		started := filled.Dx() == 0 && filled.Dy() == 0
		m.buttons.List[0].Hidden = started
		m.buttons.List[1].Hidden = started
		m.buttons.Active = 0
	case tea.KeyMsg:
		m.error = nil
		switch {
		case key.Matches(msg, m.keymap.up):
			m.buttons.Update(buttons.MoveUp)
		case key.Matches(msg, m.keymap.down):
			m.buttons.Update(buttons.MoveDown)
		case key.Matches(msg, m.keymap.choose):
			defer func() {
				m.buttons.Active = 0
			}()
			switch m.buttons.Current().Name {
			case BtnResume:
				return m, commands.ChangeView(commands.Conway)
			case BtnReset:
				m.conway.Reset()
				return m, commands.ChangeView(commands.Conway)
			case BtnNew:
				m.conway.Clear()
				return m, commands.ChangeView(commands.Conway)
			case BtnLoad:
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
				return m, m.initForm()
			case BtnQuit:
				return m, tea.Quit
			}
		case key.Matches(msg, m.keymap.resume):
			return m, commands.ChangeView(commands.Conway)
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

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

func (m *Menu) View() string {
	views := []string{Title}

	if m.form == nil {
		if p := m.conway.Pattern; p.Name != "" {
			view := "Current pattern: " + p.Name
			if p.Author != "" {
				view += " by " + p.Author
			}
			views = append(views, view+"\n")
		}

		if m.error != nil {
			views = append(views, m.styles.errorStyle.Render(m.error.Error()))
		}
		views = append(views,
			m.buttons.View()+"\n",
			m.help.ShortHelpView(m.keymap.ShortHelp()),
		)
	} else {
		views = append(views, m.form.View())
	}

	return lipgloss.Place(m.size.Width, m.size.Height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, views...),
	)
}
