package logseq

import (
	"bufio"
	"glog/domain"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// scheduledRegex matches Logseq SCHEDULED format: SCHEDULED: <2024-01-20 Sat>
var scheduledRegex = regexp.MustCompile(`SCHEDULED:\s*<(\d{4}-\d{2}-\d{2})(?:\s+\w+)?>`)

// ParseJournalFilename extracts the date from a Logseq journal filename.
// Logseq journal files are named YYYY_MM_DD.md
func ParseJournalFilename(filename string) (time.Time, error) {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Parse date from YYYY_MM_DD format
	return time.Parse("2006_01_02", name)
}

// ParsePageFilename extracts the title from a Logseq page filename.
// Handles URL-encoded filenames (e.g., "Project%20Notes.md" -> "Project Notes")
func ParsePageFilename(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// URL decode the filename (Logseq URL-encodes special characters)
	decoded, err := url.QueryUnescape(name)
	if err != nil {
		// If decoding fails, use the original name
		return name
	}

	return decoded
}

// JournalTitleFromDate formats a date into glog's journal title format.
// Example: "Monday, January 2, 2006"
func JournalTitleFromDate(date time.Time) string {
	return date.Format("Monday, January 2, 2006")
}

// ConvertScheduledInLine converts Logseq SCHEDULED format to glog format within a line.
// Input:  "Task to do SCHEDULED: <2024-01-20 Sat>"
// Output: "Task to do /scheduled 2024-01-20"
func ConvertScheduledInLine(line string) string {
	return scheduledRegex.ReplaceAllString(line, "/scheduled $1")
}

// ParseContent parses Logseq markdown content into glog blocks.
// Handles bullet-point indentation and converts SCHEDULED format.
func ParseContent(content string) []*domain.Block {
	var blocks []*domain.Block
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentBlock *domain.Block
	var continuationIndent int
	var inCodeBlock bool  // Track whether we're inside a fenced code block
	var hadCodeBlock bool // Track if current block has had a code block (for post-code-block formatting)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines at the start
		if len(blocks) == 0 && currentBlock == nil && strings.TrimSpace(line) == "" {
			continue
		}

		// Check if this is a bullet point line
		indent, content, isBullet := parseBulletLine(line)

		if isBullet {
			// Save the previous block if exists
			if currentBlock != nil {
				currentBlock.Content = strings.TrimSpace(currentBlock.Content)
				currentBlock.Content = ConvertScheduledInLine(currentBlock.Content)
				blocks = append(blocks, currentBlock)
			}

			// Reset code block state for new block
			inCodeBlock = false
			hadCodeBlock = false

			// Start new block
			currentBlock = &domain.Block{
				ID:      domain.BlockID(uuid.New()),
				Content: content,
				Indent:  indent,
			}
			continuationIndent = indent

			// Check if bullet content itself starts a code fence
			if strings.HasPrefix(strings.TrimSpace(content), "```") {
				inCodeBlock = true
			}
		} else if currentBlock != nil {
			// This is a continuation line (property, SCHEDULED, or multi-line content)
			trimmed := strings.TrimSpace(line)

			// Check for code fence toggle
			if strings.HasPrefix(trimmed, "```") {
				inCodeBlock = !inCodeBlock
				if !inCodeBlock {
					// Just closed a code block
					hadCodeBlock = true
				}
			}

			if inCodeBlock || strings.HasPrefix(trimmed, "```") {
				// Inside code block or this is a fence line: preserve newlines and relative indentation
				strippedLine := stripContinuationIndent(line, continuationIndent)
				currentBlock.Content += "\n" + strippedLine
			} else if hadCodeBlock {
				// After a code block - preserve newlines to maintain formatting
				strippedLine := stripContinuationIndent(line, continuationIndent)
				if strings.TrimSpace(strippedLine) != "" {
					currentBlock.Content += "\n" + strippedLine
				}
			} else {
				// Outside code block: original space-joining behavior
				if trimmed != "" {
					if currentBlock.Content != "" {
						currentBlock.Content += " " + trimmed
					} else {
						currentBlock.Content = trimmed
					}
				}
			}
		} else if strings.TrimSpace(line) != "" {
			// Non-bullet content at the start - create a block for it
			currentBlock = &domain.Block{
				ID:      domain.BlockID(uuid.New()),
				Content: strings.TrimSpace(line),
				Indent:  0,
			}
			continuationIndent = 0
			// Check if this starts a code fence
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inCodeBlock = true
			}
		}

		// Ignore the continuationIndent variable warning - kept for potential future use
		_ = continuationIndent
	}

	// Don't forget the last block
	if currentBlock != nil {
		currentBlock.Content = strings.TrimSpace(currentBlock.Content)
		currentBlock.Content = ConvertScheduledInLine(currentBlock.Content)
		blocks = append(blocks, currentBlock)
	}

	// If no blocks were created, create an empty one
	if len(blocks) == 0 {
		blocks = append(blocks, &domain.Block{
			ID:      domain.BlockID(uuid.New()),
			Content: "",
			Indent:  0,
		})
	}

	return blocks
}

// parseBulletLine parses a line to extract indent level and content.
// Returns (indent, content, isBullet)
// Logseq uses "- " for bullets with tabs or spaces for indentation.
func parseBulletLine(line string) (int, string, bool) {
	// Count leading whitespace
	indent := 0
	i := 0

	for i < len(line) {
		if line[i] == '\t' {
			indent++
			i++
		} else if line[i] == ' ' {
			// Count spaces - typically 2 spaces = 1 indent level in Logseq
			spaceCount := 0
			for i < len(line) && line[i] == ' ' {
				spaceCount++
				i++
			}
			indent += spaceCount / 2
			break
		} else {
			break
		}
	}

	// Check for bullet point
	remaining := line[i:]
	if strings.HasPrefix(remaining, "- ") {
		content := strings.TrimPrefix(remaining, "- ")
		return indent, content, true
	}

	// Also handle "* " bullets (some Logseq exports use this)
	if strings.HasPrefix(remaining, "* ") {
		content := strings.TrimPrefix(remaining, "* ")
		return indent, content, true
	}

	return indent, remaining, false
}

// stripContinuationIndent removes the base indentation level from a continuation line
// while preserving any additional indentation (e.g., for code inside code blocks).
// In Logseq, continuation lines are indented at (baseIndent * 2 + 2) spaces.
func stripContinuationIndent(line string, baseIndent int) string {
	// Continuation lines are indented 2 spaces beyond the bullet's indentation
	// Each indent level = 2 spaces, plus 2 spaces for content continuation
	spacesToRemove := (baseIndent * 2) + 2

	i := 0
	removed := 0
	for i < len(line) && removed < spacesToRemove {
		if line[i] == ' ' {
			removed++
			i++
		} else if line[i] == '\t' {
			// Treat tab as equivalent to reaching an indent level
			removed += 2
			i++
		} else {
			break
		}
	}
	return line[i:]
}

// ParseFile parses a Logseq file into a domain.Document.
func ParseFile(path string, isJournal bool) (*domain.Document, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(path)

	var title string
	var date time.Time

	if isJournal {
		// Parse date from filename
		date, err = ParseJournalFilename(filename)
		if err != nil {
			return nil, err
		}
		title = JournalTitleFromDate(date)
	} else {
		// Page - title from filename
		title = ParsePageFilename(filename)
		date = time.Now()
	}

	// Parse content into blocks
	blocks := ParseContent(string(content))

	doc := &domain.Document{
		ID:        domain.DocumentID(uuid.New()),
		Title:     title,
		Date:      date,
		IsJournal: isJournal,
		Blocks:    blocks,
	}

	return doc, nil
}
