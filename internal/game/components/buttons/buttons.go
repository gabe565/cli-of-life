package buttons

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Button struct {
	Name   string
	Hidden bool
}

func New(names ...string) *Buttons {
	btns := make([]*Button, 0, len(names))
	for _, name := range names {
		btns = append(btns, &Button{Name: name})
	}

	bgColor := lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#4A4A4A"}

	btnStyle := lipgloss.NewStyle().
		Border(lipgloss.InnerHalfBlockBorder()).
		BorderForeground(bgColor).
		Padding(0, 1).
		Background(bgColor).
		Width(20)

	selectedBgColor := lipgloss.AdaptiveColor{Light: "#aaf", Dark: "#4A4ABB"}

	return &Buttons{
		List: btns,
		styles: styles{
			button: btnStyle,
			selected: btnStyle.Bold(true).
				Background(selectedBgColor).
				BorderForeground(selectedBgColor),
		},
		Position: lipgloss.Center,
	}
}

type styles struct {
	button   lipgloss.Style
	selected lipgloss.Style
}

type Buttons struct {
	styles styles

	Position lipgloss.Position
	List     []*Button
	Active   int
}

func (b *Buttons) Init() tea.Cmd {
	return nil
}

func (b *Buttons) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) { //nolint:gocritic
	case Move:
		switch msg {
		case MoveUp:
			if b.Active > 0 {
				b.Active--
				if b.List[b.Active].Hidden {
					return b.Update(msg)
				}
			}
		case MoveDown:
			if b.Active < len(b.List)-1 {
				b.Active++
				if b.List[b.Active].Hidden {
					return b.Update(msg)
				}
			}
		}
	}
	return b, nil
}

func (b *Buttons) Current() *Button {
	return b.List[b.Active]
}

type Move uint8

const (
	MoveUp Move = iota
	MoveDown
)

func (b *Buttons) View() string {
	fields := make([]string, 0, len(b.List))
	for i, btn := range b.List {
		var view string
		if i == b.Active {
			if btn.Hidden && b.Active < len(b.List)-1 {
				b.Active++
			}
			view = b.styles.selected.Render(btn.Name)
		} else {
			view = b.styles.button.Render(btn.Name)
		}
		if !btn.Hidden {
			fields = append(fields, view)
		}
	}
	return lipgloss.JoinVertical(b.Position, fields...)
}
