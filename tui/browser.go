package tui

import (
	"fmt"
	"strings"
	"time"

	"glog/domain"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewMode int

const (
	modeList viewMode = iota
	modeDetail
)

type BrowserModel struct {
	store         *domain.DocumentStore
	documents     []domain.DocumentID
	docMetadata   map[domain.DocumentID]*domain.Document
	currentIndex  int
	viewMode      viewMode
	selectedDoc   *domain.Document
	selectedParas map[domain.ParagraphID]domain.Paragraph
	err           error
	loading       bool
	width         int
	height        int
}

type documentsLoadedMsg struct {
	docs     []domain.DocumentID
	metadata map[domain.DocumentID]*domain.Document
}

type documentDetailLoadedMsg struct {
	doc   *domain.Document
	paras map[domain.ParagraphID]domain.Paragraph
}

var (
	listHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	listItemStyle = lipgloss.NewStyle().
			Padding(0, 2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("99")).
				Bold(true).
				Padding(0, 2)

	docTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Underline(true)

	docMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true)

	docContentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(1, 2)
)

func NewBrowserModel(store *domain.DocumentStore) BrowserModel {
	return BrowserModel{
		store:        store,
		documents:    []domain.DocumentID{},
		docMetadata:  make(map[domain.DocumentID]*domain.Document),
		currentIndex: 0,
		viewMode:     modeList,
		loading:      true,
	}
}

func (m BrowserModel) Init() tea.Cmd {
	return m.loadDocuments()
}

func (m BrowserModel) loadDocuments() tea.Cmd {
	return func() tea.Msg {
		// Load documents from the last 10 days
		now := time.Now()
		from := domain.ToDateTime(now.AddDate(0, 0, -10))
		to := domain.ToDateTime(now.AddDate(0, 0, 1)) // Include today + 1 day buffer

		docIDs, err := m.store.Load(from, to)
		if err != nil {
			return errMsg{err: err}
		}

		// Load metadata for each document
		metadata := make(map[domain.DocumentID]*domain.Document)
		for _, docID := range docIDs {
			doc, _, err := m.store.LoadDocument(docID)
			if err != nil {
				continue // Skip documents that fail to load
			}
			metadata[docID] = doc
		}

		return documentsLoadedMsg{
			docs:     docIDs,
			metadata: metadata,
		}
	}
}

func (m BrowserModel) loadDocumentDetail() tea.Cmd {
	if len(m.documents) == 0 {
		return nil
	}

	docID := m.documents[m.currentIndex]
	return func() tea.Msg {
		doc, paras, err := m.store.LoadDocument(docID)
		if err != nil {
			return errMsg{err: err}
		}

		return documentDetailLoadedMsg{
			doc:   doc,
			paras: paras,
		}
	}
}

func (m BrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case documentsLoadedMsg:
		m.documents = msg.docs
		m.docMetadata = msg.metadata
		m.loading = false
		if len(m.documents) > 0 {
			m.currentIndex = 0
		}
		return m, nil

	case documentDetailLoadedMsg:
		m.selectedDoc = msg.doc
		m.selectedParas = msg.paras
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.viewMode == modeList && len(m.documents) > 0 {
				if m.currentIndex > 0 {
					m.currentIndex--
				}
			}

		case "down", "j":
			if m.viewMode == modeList && len(m.documents) > 0 {
				if m.currentIndex < len(m.documents)-1 {
					m.currentIndex++
				}
			}

		case "enter":
			if m.viewMode == modeList && len(m.documents) > 0 {
				m.viewMode = modeDetail
				return m, m.loadDocumentDetail()
			}

		case "backspace", "left", "h":
			if m.viewMode == modeDetail {
				m.viewMode = modeList
				m.selectedDoc = nil
				m.selectedParas = nil
			}

		case "r":
			// Refresh the list
			m.loading = true
			return m, m.loadDocuments()
		}
	}

	return m, nil
}

func (m BrowserModel) View() string {
	if m.loading {
		return "\n  Loading documents from the last 10 days...\n\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n\n  Press 'q' or ESC to quit\n", m.err))
	}

	if m.viewMode == modeDetail {
		return m.viewDetail()
	}

	return m.viewList()
}

func (m BrowserModel) viewList() string {
	var b strings.Builder

	// Header
	b.WriteString(listHeaderStyle.Render(" 📚 Documents (Last 10 Days) "))
	b.WriteString("\n\n")

	if len(m.documents) == 0 {
		b.WriteString("  No documents found.\n\n")
		b.WriteString(helpStyle.Render("  Press 'r' to refresh | 'q' or ESC to quit"))
		return b.String()
	}

	// Document list
	visibleStart := 0
	visibleEnd := len(m.documents)

	// Limit visible items if we have many documents
	maxVisible := 15
	if len(m.documents) > maxVisible {
		// Center the current selection
		visibleStart = m.currentIndex - maxVisible/2
		if visibleStart < 0 {
			visibleStart = 0
		}
		visibleEnd = visibleStart + maxVisible
		if visibleEnd > len(m.documents) {
			visibleEnd = len(m.documents)
			visibleStart = visibleEnd - maxVisible
			if visibleStart < 0 {
				visibleStart = 0
			}
		}
	}

	for i := visibleStart; i < visibleEnd; i++ {
		docID := m.documents[i]
		doc := m.docMetadata[docID]

		if doc == nil {
			continue
		}

		// Format the date
		docTime, _ := doc.Date.ToTime()
		dateStr := docTime.Format("Jan 02, 2006 15:04")

		line := fmt.Sprintf("%s  %s", dateStr, doc.Title)

		if i == m.currentIndex {
			b.WriteString(selectedItemStyle.Render("▶ " + line))
		} else {
			b.WriteString(listItemStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}

	// Show indicator if there are more items
	if visibleStart > 0 {
		b.WriteString(helpStyle.Render("\n  ↑ More items above..."))
	}
	if visibleEnd < len(m.documents) {
		b.WriteString(helpStyle.Render("\n  ↓ More items below..."))
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf(
		"  ↑/↓ or j/k: Navigate | Enter: View detail | r: Refresh | q/ESC: Quit  [%d/%d]",
		m.currentIndex+1,
		len(m.documents),
	)))

	return b.String()
}

func (m BrowserModel) viewDetail() string {
	if m.selectedDoc == nil {
		return "\n  Loading document details...\n\n"
	}

	var b strings.Builder

	// Title
	b.WriteString(docTitleStyle.Render(m.selectedDoc.Title))
	b.WriteString("\n")

	// Metadata
	docTime, _ := m.selectedDoc.Date.ToTime()
	dateStr := docTime.Format("January 02, 2006 at 15:04")
	b.WriteString(docMetaStyle.Render(fmt.Sprintf("Created: %s | ID: %s", dateStr, m.selectedDoc.ID.String()[:8])))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 80))
	b.WriteString("\n\n")

	// Content (paragraphs)
	for _, paraID := range m.selectedDoc.Body {
		para, exists := m.selectedParas[paraID]
		if !exists {
			continue
		}

		content := string(para.Content)
		b.WriteString(docContentStyle.Render(content))
		b.WriteString("\n\n")
	}

	b.WriteString(strings.Repeat("─", 80))
	b.WriteString("\n")

	// Help text
	b.WriteString(helpStyle.Render("  Backspace/Left/h: Back to list | q/ESC: Quit"))

	return b.String()
}

func RunBrowser(store *domain.DocumentStore) error {
	p := tea.NewProgram(NewBrowserModel(store))
	_, err := p.Run()
	return err
}
