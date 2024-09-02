package menu

import (
	"errors"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/game/commands"
	"github.com/gabe565/cli-of-life/internal/game/components/buttons"
	"github.com/gabe565/cli-of-life/internal/game/conway"
	"github.com/gabe565/cli-of-life/internal/pattern"
	zone "github.com/lrstanley/bubblezone"
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
	zone.NewGlobal()
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

	wasPressed bool
	conway     *conway.Conway
	buttons    *buttons.Buttons
	form       *huh.Form
	patternSrc string

	error error
}

func (m *Menu) Init() tea.Cmd {
	if m.config.Pattern != "" {
		return m.LoadPattern()
	}
	return nil
}

func (m *Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = msg
		m.styles.errorStyle = m.styles.errorStyle.Width(msg.Width)
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
			defer func() {
				m.patternSrc = ""
			}()
			switch m.patternSrc {
			case sourceEmbedded:
				return m, m.patternEmbeddedForm()
			case sourceFile:
				return m, m.patternFileForm()
			case sourceURL:
				return m, m.patternURLForm()
			default:
				m.conway.ResumeOnFocus = false
				return m, m.LoadPattern()
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
		if msg == commands.Menu {
			empty := m.conway.Pattern.Tree.IsEmpty()
			m.buttons.List[0].Hidden = empty
			m.buttons.List[1].Hidden = empty
			m.buttons.Active = 0
		}
	case tea.KeyMsg:
		m.error = nil
		switch {
		case key.Matches(msg, m.keymap.up):
			m.buttons.Update(buttons.MoveUp)
		case key.Matches(msg, m.keymap.down):
			m.buttons.Update(buttons.MoveDown)
		case key.Matches(msg, m.keymap.choose):
			return m, m.handleButtonPress(m.buttons.Current())
		case key.Matches(msg, m.keymap.resume):
			return m, commands.ChangeView(commands.Conway)
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		}
	case tea.MouseMsg:
		defer func() {
			m.wasPressed = msg.Action == tea.MouseActionPress
		}()
		switch msg.Action {
		case tea.MouseActionRelease, tea.MouseActionMotion:
			for i, btn := range m.buttons.List {
				if zone.Get(btn.ID()).InBounds(msg) {
					if m.wasPressed && msg.Action == tea.MouseActionRelease {
						return m, m.handleButtonPress(btn)
					}
					m.buttons.Active = i
					break
				}
			}
		}
	}
	return m, nil
}

func (m *Menu) LoadPattern() tea.Cmd {
	p, err := pattern.New(m.config)
	if err != nil {
		var multiplePatterns pattern.MultiplePatternsError
		if errors.As(err, &multiplePatterns) {
			return m.choosePatternForm(multiplePatterns.URLs)
		}
		m.error = err
		return nil
	}
	m.conway.Pattern = p
	m.conway.ResetView()
	return commands.ChangeView(commands.Conway)
}

func (m *Menu) View() string {
	views := []string{Title}

	if m.form == nil {
		if m.conway.Pattern != nil {
			if p := m.conway.Pattern; p.Name != "" {
				views = append(views, "Current pattern: "+p.NameAuthor()+"\n")
			}
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

	return zone.Scan(lipgloss.Place(m.size.Width, m.size.Height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, views...),
	))
}
