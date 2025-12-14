package tui

import (
	"fmt"
	"strings"
	"time"

	"glog/domain"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type focusField int

const (
	focusTitle focusField = iota
	focusContent
	focusButtons
)

type Model struct {
	store         *domain.DocumentStore
	titleInput    textinput.Model
	contentInput  textarea.Model
	focusedField  focusField
	buttonFocused int // 0 = Save, 1 = Cancel
	err           error
	saved         bool
	width         int
	height        int
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	buttonStyle = lipgloss.NewStyle().
			Padding(0, 3).
			MarginRight(2)

	activeButtonStyle = buttonStyle.
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("205")).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
)

func NewModel(store *domain.DocumentStore) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter document title..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 80

	ta := textarea.New()
	ta.Placeholder = "Enter document content..."
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.ShowLineNumbers = false

	return Model{
		store:         store,
		titleInput:    ti,
		contentInput:  ta,
		focusedField:  focusTitle,
		buttonFocused: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case errMsg:
		m.err = msg.err
		return m, nil

	case savedMsg:
		m.saved = true
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab":
			if msg.String() == "tab" {
				m.focusedField = focusField((int(m.focusedField) + 1) % 3)
			} else {
				m.focusedField = focusField((int(m.focusedField) - 1 + 3) % 3)
			}

			if m.focusedField == focusTitle {
				m.titleInput.Focus()
				m.contentInput.Blur()
			} else if m.focusedField == focusContent {
				m.titleInput.Blur()
				m.contentInput.Focus()
			} else {
				m.titleInput.Blur()
				m.contentInput.Blur()
			}
			return m, nil

		case "enter":
			if m.focusedField == focusButtons {
				if m.buttonFocused == 0 {
					// Save button
					return m, m.saveDocument()
				} else {
					// Cancel button
					return m, tea.Quit
				}
			} else if m.focusedField == focusContent {
				// Allow Enter in content area for new lines
				m.contentInput, cmd = m.contentInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case "left", "right":
			if m.focusedField == focusButtons {
				if msg.String() == "left" {
					m.buttonFocused = 0
				} else {
					m.buttonFocused = 1
				}
			}

		default:
			if m.focusedField == focusTitle {
				m.titleInput, cmd = m.titleInput.Update(msg)
				cmds = append(cmds, cmd)
			} else if m.focusedField == focusContent {
				m.contentInput, cmd = m.contentInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) saveDocument() tea.Cmd {
	return func() tea.Msg {
		title := strings.TrimSpace(m.titleInput.Value())
		content := strings.TrimSpace(m.contentInput.Value())

		if title == "" {
			return errMsg{err: fmt.Errorf("title cannot be empty")}
		}

		if content == "" {
			return errMsg{err: fmt.Errorf("content cannot be empty")}
		}

		// Create paragraph
		paraID := domain.ParagraphID(uuid.New())
		docID := domain.DocumentID(uuid.New())

		para := domain.Paragraph{
			ID:         paraID,
			DocumentID: docID,
			Parent:     nil,
			Children:   nil,
			Content:    domain.Content(content),
			References: []domain.ParagraphID{},
		}

		// Create document
		doc := &domain.Document{
			ID:    docID,
			Title: title,
			Date:  domain.ToDateTime(time.Now()),
			Body:  []domain.ParagraphID{paraID},
		}

		// Save to store
		err := m.store.Save(doc, []domain.Paragraph{para})
		if err != nil {
			return errMsg{err: err}
		}

		return savedMsg{}
	}
}

type errMsg struct {
	err error
}

type savedMsg struct{}

func (m Model) View() string {
	if m.saved {
		return successStyle.Render("✓ Document saved successfully!\n\nPress ESC or Ctrl+C to exit.")
	}

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("📝 Create New Document"))
	b.WriteString("\n\n")

	// Title input
	titleLabel := "Title:"
	if m.focusedField == focusTitle {
		titleLabel = focusedStyle.Render("▶ Title:")
	} else {
		titleLabel = blurredStyle.Render("  Title:")
	}
	b.WriteString(titleLabel)
	b.WriteString("\n")
	b.WriteString(m.titleInput.View())
	b.WriteString("\n\n")

	// Content input
	contentLabel := "Content:"
	if m.focusedField == focusContent {
		contentLabel = focusedStyle.Render("▶ Content:")
	} else {
		contentLabel = blurredStyle.Render("  Content:")
	}
	b.WriteString(contentLabel)
	b.WriteString("\n")
	b.WriteString(m.contentInput.View())
	b.WriteString("\n\n")

	// Buttons
	saveBtn := "Save"
	cancelBtn := "Cancel"

	if m.focusedField == focusButtons {
		if m.buttonFocused == 0 {
			saveBtn = activeButtonStyle.Render(saveBtn)
			cancelBtn = buttonStyle.Render(cancelBtn)
		} else {
			saveBtn = buttonStyle.Render(saveBtn)
			cancelBtn = activeButtonStyle.Render(cancelBtn)
		}
	} else {
		saveBtn = buttonStyle.Render(saveBtn)
		cancelBtn = buttonStyle.Render(cancelBtn)
	}

	b.WriteString(saveBtn)
	b.WriteString(cancelBtn)
	b.WriteString("\n")

	// Error message
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString(helpStyle.Render("\nTab/Shift+Tab: Navigate | Enter: Submit | Left/Right: Switch buttons | ESC: Quit"))

	return b.String()
}

func (m Model) updateFromMsg(msg tea.Msg) Model {
	switch msg := msg.(type) {
	case errMsg:
		m.err = msg.err
	case savedMsg:
		m.saved = true
	}
	return m
}

func Run(store *domain.DocumentStore) error {
	p := tea.NewProgram(NewModel(store))
	_, err := p.Run()
	return err
}
