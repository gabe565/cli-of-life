package menu

import (
	"errors"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"gabe565.com/cli-of-life/internal/config"
	"gabe565.com/cli-of-life/internal/game/commands"
	"gabe565.com/cli-of-life/internal/game/components/buttons"
	"gabe565.com/cli-of-life/internal/game/conway"
	"gabe565.com/cli-of-life/internal/pattern"
	"gabe565.com/cli-of-life/internal/quadtree"
	zone "github.com/lrstanley/bubblezone/v2"
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
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
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
	case commands.ViewMsg:
		if msg == commands.Menu {
			empty := m.conway.Pattern.Tree.IsEmpty()
			m.buttons.List[0].Hidden = empty
			m.buttons.List[1].Hidden = empty
			m.buttons.Active = 0
		}
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.up):
			m.buttons.Update(buttons.MoveUp)
		case key.Matches(msg, m.keymap.down):
			m.buttons.Update(buttons.MoveDown)
		case key.Matches(msg, m.keymap.choose):
			m.error = nil
			return m, m.handleButtonPress(m.buttons.Current())
		case key.Matches(msg, m.keymap.resume):
			m.error = nil
			return m, commands.ChangeView(commands.Conway)
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		}
	case tea.MouseMsg:
		_, isRelease := msg.(tea.MouseReleaseMsg)
		_, isMotion := msg.(tea.MouseMotionMsg)
		defer func() {
			_, m.wasPressed = msg.(tea.MouseClickMsg)
		}()
		if isRelease || isMotion {
			for i, btn := range m.buttons.List {
				if zone.Get(btn.ID()).InBounds(msg) {
					if m.wasPressed && isRelease {
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
	quadtree.ResetCache()
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

func (m *Menu) SetDark(dark bool) {
	m.help.Styles = help.DefaultStyles(dark)
	m.buttons.SetDark(dark)
}

func (m *Menu) View() tea.View {
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

	return tea.NewView(zone.Scan(lipgloss.Place(m.size.Width, m.size.Height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, views...),
	)))
}
