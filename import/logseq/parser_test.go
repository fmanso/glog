package logseq

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseJournalFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantDate time.Time
		wantErr  bool
	}{
		{
			name:     "valid date",
			filename: "2024_01_15.md",
			wantDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "end of year",
			filename: "2023_12_31.md",
			wantDate: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "start of year",
			filename: "2024_01_01.md",
			wantDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "invalid format",
			filename: "2024-01-15.md",
			wantErr:  true,
		},
		{
			name:     "not a date",
			filename: "notes.md",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJournalFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJournalFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.wantDate) {
				t.Errorf("ParseJournalFilename() = %v, want %v", got, tt.wantDate)
			}
		})
	}
}

func TestParsePageFilename(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		wantTitle string
	}{
		{
			name:      "simple filename",
			filename:  "My Page.md",
			wantTitle: "My Page",
		},
		{
			name:      "URL encoded spaces",
			filename:  "Project%20Notes.md",
			wantTitle: "Project Notes",
		},
		{
			name:      "URL encoded special chars",
			filename:  "C%2B%2B%20Programming.md",
			wantTitle: "C++ Programming",
		},
		{
			name:      "no extension edge case",
			filename:  "NoExtension",
			wantTitle: "NoExtension",
		},
		{
			name:      "multiple dots",
			filename:  "file.name.with.dots.md",
			wantTitle: "file.name.with.dots",
		},
		{
			name:      "URL encoded slash",
			filename:  "Project%2FNotes.md",
			wantTitle: "Project/Notes",
		},
		{
			name:      "multiple slashes",
			filename:  "Work%2FProjects%2FAlpha.md",
			wantTitle: "Work/Projects/Alpha",
		},
		{
			name:      "slash with spaces",
			filename:  "My%20Project%2FSubfolder.md",
			wantTitle: "My Project/Subfolder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePageFilename(tt.filename)
			if got != tt.wantTitle {
				t.Errorf("ParsePageFilename() = %v, want %v", got, tt.wantTitle)
			}
		})
	}
}

func TestJournalTitleFromDate(t *testing.T) {
	tests := []struct {
		name      string
		date      time.Time
		wantTitle string
	}{
		{
			name:      "typical date",
			date:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantTitle: "Monday, January 15, 2024",
		},
		{
			name:      "single digit day",
			date:      time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
			wantTitle: "Tuesday, March 5, 2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JournalTitleFromDate(tt.date)
			if got != tt.wantTitle {
				t.Errorf("JournalTitleFromDate() = %v, want %v", got, tt.wantTitle)
			}
		})
	}
}

func TestConvertScheduledInLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "standard scheduled",
			input: "Task to do SCHEDULED: <2024-01-20 Sat>",
			want:  "Task to do /scheduled 2024-01-20",
		},
		{
			name:  "scheduled without day name",
			input: "Task SCHEDULED: <2024-01-20>",
			want:  "Task /scheduled 2024-01-20",
		},
		{
			name:  "no scheduled",
			input: "Regular content without scheduled",
			want:  "Regular content without scheduled",
		},
		{
			name:  "scheduled at start",
			input: "SCHEDULED: <2024-12-31 Mon> End of year",
			want:  "/scheduled 2024-12-31 End of year",
		},
		{
			name:  "multiple scheduled on line",
			input: "SCHEDULED: <2024-01-20 Sat> and SCHEDULED: <2024-01-21 Sun>",
			want:  "/scheduled 2024-01-20 and /scheduled 2024-01-21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertScheduledInLine(tt.input)
			if got != tt.want {
				t.Errorf("ConvertScheduledInLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseContent(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantBlocks int
		checkFirst string
		checkLast  string
	}{
		{
			name: "simple bullets",
			content: `- First block
- Second block
- Third block`,
			wantBlocks: 3,
			checkFirst: "First block",
			checkLast:  "Third block",
		},
		{
			name: "nested bullets",
			content: `- Parent
  - Child
    - Grandchild`,
			wantBlocks: 3,
			checkFirst: "Parent",
			checkLast:  "Grandchild",
		},
		{
			name: "with scheduled",
			content: `- Task to complete
  SCHEDULED: <2024-01-20 Sat>`,
			wantBlocks: 1,
			checkFirst: "Task to complete /scheduled 2024-01-20",
		},
		{
			name: "with wikilinks",
			content: `- Link to [[Another Page]]
- Also references [[Project A]]`,
			wantBlocks: 2,
			checkFirst: "Link to [[Another Page]]",
			checkLast:  "Also references [[Project A]]",
		},
		{
			name: "wikilinks with slashes",
			content: `- See [[Project/Notes]] for details
- Check [[Work/Projects/Alpha]]`,
			wantBlocks: 2,
			checkFirst: "See [[Project/Notes]] for details",
			checkLast:  "Check [[Work/Projects/Alpha]]",
		},
		{
			name: "with properties",
			content: `- Block with property
  tags:: programming, golang`,
			wantBlocks: 1,
			checkFirst: "Block with property tags:: programming, golang",
		},
		{
			name:       "empty content",
			content:    "",
			wantBlocks: 1,
			checkFirst: "",
		},
		{
			name: "leading empty lines",
			content: `

- First actual block`,
			wantBlocks: 1,
			checkFirst: "First actual block",
		},
		{
			name: "tab indentation",
			content: `- Root
	- Tab indented child`,
			wantBlocks: 2,
			checkFirst: "Root",
			checkLast:  "Tab indented child",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := ParseContent(tt.content)

			if len(blocks) != tt.wantBlocks {
				t.Errorf("ParseContent() returned %d blocks, want %d", len(blocks), tt.wantBlocks)
				for i, b := range blocks {
					t.Logf("  Block %d: indent=%d content=%q", i, b.Indent, b.Content)
				}
				return
			}

			if tt.checkFirst != "" && blocks[0].Content != tt.checkFirst {
				t.Errorf("First block content = %q, want %q", blocks[0].Content, tt.checkFirst)
			}

			if tt.checkLast != "" && len(blocks) > 0 {
				lastBlock := blocks[len(blocks)-1]
				if lastBlock.Content != tt.checkLast {
					t.Errorf("Last block content = %q, want %q", lastBlock.Content, tt.checkLast)
				}
			}
		})
	}
}

func TestParseContentIndentation(t *testing.T) {
	content := `- Root level
  - First indent
    - Second indent
  - Back to first
- Back to root`

	blocks := ParseContent(content)

	expectedIndents := []int{0, 1, 2, 1, 0}

	if len(blocks) != len(expectedIndents) {
		t.Fatalf("Expected %d blocks, got %d", len(expectedIndents), len(blocks))
	}

	for i, block := range blocks {
		if block.Indent != expectedIndents[i] {
			t.Errorf("Block %d: indent = %d, want %d (content: %q)", i, block.Indent, expectedIndents[i], block.Content)
		}
	}
}

func TestParseFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "logseq-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("parse journal file", func(t *testing.T) {
		// Create a test journal file
		journalContent := `- Morning standup
  - Discussed [[Project Alpha]]
- Read article about Go
  tags:: programming`

		journalPath := filepath.Join(tmpDir, "2024_01_15.md")
		if err := os.WriteFile(journalPath, []byte(journalContent), 0644); err != nil {
			t.Fatal(err)
		}

		doc, err := ParseFile(journalPath, true)
		if err != nil {
			t.Fatalf("ParseFile() error = %v", err)
		}

		if !doc.IsJournal {
			t.Error("Expected IsJournal to be true")
		}

		expectedTitle := "Monday, January 15, 2024"
		if doc.Title != expectedTitle {
			t.Errorf("Title = %q, want %q", doc.Title, expectedTitle)
		}

		if len(doc.Blocks) != 3 {
			t.Errorf("Expected 3 blocks, got %d", len(doc.Blocks))
		}
	})

	t.Run("parse page file", func(t *testing.T) {
		// Create a test page file
		pageContent := `- This is a note about Project Alpha
- Related to [[Other Project]]`

		pagePath := filepath.Join(tmpDir, "Project%20Alpha.md")
		if err := os.WriteFile(pagePath, []byte(pageContent), 0644); err != nil {
			t.Fatal(err)
		}

		doc, err := ParseFile(pagePath, false)
		if err != nil {
			t.Fatalf("ParseFile() error = %v", err)
		}

		if doc.IsJournal {
			t.Error("Expected IsJournal to be false")
		}

		expectedTitle := "Project Alpha"
		if doc.Title != expectedTitle {
			t.Errorf("Title = %q, want %q", doc.Title, expectedTitle)
		}

		if len(doc.Blocks) != 2 {
			t.Errorf("Expected 2 blocks, got %d", len(doc.Blocks))
		}
	})
}

func TestParseBulletLine(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantIndent int
		wantText   string
		wantBullet bool
	}{
		{
			name:       "root bullet",
			line:       "- Content",
			wantIndent: 0,
			wantText:   "Content",
			wantBullet: true,
		},
		{
			name:       "two space indent",
			line:       "  - Indented",
			wantIndent: 1,
			wantText:   "Indented",
			wantBullet: true,
		},
		{
			name:       "four space indent",
			line:       "    - Double indent",
			wantIndent: 2,
			wantText:   "Double indent",
			wantBullet: true,
		},
		{
			name:       "tab indent",
			line:       "\t- Tab indented",
			wantIndent: 1,
			wantText:   "Tab indented",
			wantBullet: true,
		},
		{
			name:       "not a bullet",
			line:       "Regular text",
			wantIndent: 0,
			wantText:   "Regular text",
			wantBullet: false,
		},
		{
			name:       "asterisk bullet",
			line:       "* Star bullet",
			wantIndent: 0,
			wantText:   "Star bullet",
			wantBullet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indent, text, isBullet := parseBulletLine(tt.line)

			if indent != tt.wantIndent {
				t.Errorf("indent = %d, want %d", indent, tt.wantIndent)
			}
			if text != tt.wantText {
				t.Errorf("text = %q, want %q", text, tt.wantText)
			}
			if isBullet != tt.wantBullet {
				t.Errorf("isBullet = %v, want %v", isBullet, tt.wantBullet)
			}
		})
	}
}

func TestParseContentCodeBlocks(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantBlocks  int
		wantContent string
	}{
		{
			name: "code block preserves newlines",
			content: "- Here is code:\n" +
				"  ```go\n" +
				"  func main() {\n" +
				"      fmt.Println(\"Hello\")\n" +
				"  }\n" +
				"  ```",
			wantBlocks:  1,
			wantContent: "Here is code:\n```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
		},
		{
			name: "code block with empty lines",
			content: "- Code:\n" +
				"  ```\n" +
				"  line1\n" +
				"  \n" +
				"  line3\n" +
				"  ```",
			wantBlocks:  1,
			wantContent: "Code:\n```\nline1\n\nline3\n```",
		},
		{
			name: "bullet starting with code fence",
			content: "- ```js\n" +
				"  console.log('hi')\n" +
				"  ```",
			wantBlocks:  1,
			wantContent: "```js\nconsole.log('hi')\n```",
		},
		{
			name: "multiple blocks one with code",
			content: "- First block\n" +
				"- Code block:\n" +
				"  ```\n" +
				"  code here\n" +
				"  ```\n" +
				"- Third block",
			wantBlocks:  3,
			wantContent: "First block", // first block content
		},
		{
			name:        "inline code not affected",
			content:     "- Use `fmt.Println()` for output",
			wantBlocks:  1,
			wantContent: "Use `fmt.Println()` for output",
		},
		{
			name: "code block at nested indent",
			content: "- Parent\n" +
				"  - Child with code:\n" +
				"    ```python\n" +
				"    print('hello')\n" +
				"    ```",
			wantBlocks:  2,
			wantContent: "Parent", // first block
		},
		{
			name: "text after code block",
			content: "- Block with code:\n" +
				"  ```\n" +
				"  code\n" +
				"  ```\n" +
				"  And some text after",
			wantBlocks:  1,
			wantContent: "Block with code:\n```\ncode\n```\nAnd some text after",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := ParseContent(tt.content)

			if len(blocks) != tt.wantBlocks {
				t.Errorf("ParseContent() returned %d blocks, want %d", len(blocks), tt.wantBlocks)
				for i, b := range blocks {
					t.Logf("  Block %d: indent=%d content=%q", i, b.Indent, b.Content)
				}
				return
			}

			if tt.wantContent != "" && blocks[0].Content != tt.wantContent {
				t.Errorf("Block content = %q, want %q", blocks[0].Content, tt.wantContent)
			}
		})
	}
}

func TestParseContentCodeBlocksMultiple(t *testing.T) {
	// Test specific block contents in multi-block scenarios
	content := "- First block\n" +
		"- Code block:\n" +
		"  ```\n" +
		"  code here\n" +
		"  ```\n" +
		"- Third block"

	blocks := ParseContent(content)

	if len(blocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(blocks))
	}

	expectedContents := []string{
		"First block",
		"Code block:\n```\ncode here\n```",
		"Third block",
	}

	for i, expected := range expectedContents {
		if blocks[i].Content != expected {
			t.Errorf("Block %d content = %q, want %q", i, blocks[i].Content, expected)
		}
	}
}

func TestParseContentNestedCodeBlock(t *testing.T) {
	content := "- Parent\n" +
		"  - Child with code:\n" +
		"    ```python\n" +
		"    def hello():\n" +
		"        print('hello')\n" +
		"    ```"

	blocks := ParseContent(content)

	if len(blocks) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(blocks))
	}

	// Check first block
	if blocks[0].Content != "Parent" {
		t.Errorf("First block content = %q, want %q", blocks[0].Content, "Parent")
	}
	if blocks[0].Indent != 0 {
		t.Errorf("First block indent = %d, want 0", blocks[0].Indent)
	}

	// Check second block with code
	expectedContent := "Child with code:\n```python\ndef hello():\n    print('hello')\n```"
	if blocks[1].Content != expectedContent {
		t.Errorf("Second block content = %q, want %q", blocks[1].Content, expectedContent)
	}
	if blocks[1].Indent != 1 {
		t.Errorf("Second block indent = %d, want 1", blocks[1].Indent)
	}
}
