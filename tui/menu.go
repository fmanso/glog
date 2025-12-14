package tui

import (
	"glog/domain"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuOption int

const (
	optionCreate menuOption = iota
	optionBrowse
	optionQuit
)

type MenuModel struct {
	store          *domain.DocumentStore
	selectedOption menuOption
	width          int
	height         int
}

var (
	menuTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(2)

	menuItemStyle = lipgloss.NewStyle().
			Padding(0, 4)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("205")).
				Bold(true).
				Padding(0, 4)
)

func NewMenuModel(store *domain.DocumentStore) MenuModel {
	return MenuModel{
		store:          store,
		selectedOption: optionBrowse,
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.selectedOption > 0 {
				m.selectedOption--
			}

		case "down", "j":
			if m.selectedOption < optionQuit {
				m.selectedOption++
			}

		case "enter":
			switch m.selectedOption {
			case optionCreate:
				return NewModel(m.store), nil
			case optionBrowse:
				return NewBrowserModel(m.store), NewBrowserModel(m.store).Init()
			case optionQuit:
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m MenuModel) View() string {
	s := menuTitleStyle.Render("📝 glog - Document Management System")
	s += "\n\n"

	options := []string{
		"Create New Document",
		"Browse Documents (Last 10 Days)",
		"Quit",
	}

	for i, option := range options {
		if menuOption(i) == m.selectedOption {
			s += selectedMenuItemStyle.Render("▶ "+option) + "\n"
		} else {
			s += menuItemStyle.Render("  "+option) + "\n"
		}
	}

	s += "\n" + helpStyle.Render("  ↑/↓ or j/k: Navigate | Enter: Select | q/ESC: Quit")

	return s
}

func RunMenu(store *domain.DocumentStore) error {
	p := tea.NewProgram(NewMenuModel(store))
	_, err := p.Run()
	return err
}
