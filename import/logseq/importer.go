package logseq

import (
	"errors"
	"fmt"
	"glog/db"
	"os"
	"path/filepath"
	"strings"
)

// ImportOptions configures the import behavior.
type ImportOptions struct {
	LogseqPath   string // Path to Logseq graph directory
	DBPath       string // Path to glog database file
	JournalsOnly bool   // If true, only import journals
	PagesOnly    bool   // If true, only import pages
	DryRun       bool   // If true, don't actually write to database
	Verbose      bool   // If true, print detailed progress
}

// RenameInfo tracks when a document was renamed due to duplicate title.
type RenameInfo struct {
	OriginalTitle string
	NewTitle      string
}

// ImportResult contains the results of an import operation.
type ImportResult struct {
	JournalsImported int
	PagesImported    int
	Skipped          int
	Renamed          []RenameInfo
	Errors           []error
}

// Importer handles importing Logseq data into glog.
type Importer struct {
	opts   ImportOptions
	store  *db.DocumentStore
	result *ImportResult
}

// NewImporter creates a new Importer with the given options.
func NewImporter(opts ImportOptions) *Importer {
	return &Importer{
		opts: opts,
		result: &ImportResult{
			Renamed: make([]RenameInfo, 0),
			Errors:  make([]error, 0),
		},
	}
}

// Import performs the import operation.
func (imp *Importer) Import() (*ImportResult, error) {
	// Validate Logseq path
	if err := imp.validateLogseqPath(); err != nil {
		return nil, fmt.Errorf("invalid Logseq path: %w", err)
	}

	// Open database (unless dry run)
	if !imp.opts.DryRun {
		store, err := db.NewDocumentStore(imp.opts.DBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()
		imp.store = store
	}

	// Import journals
	if !imp.opts.PagesOnly {
		if err := imp.importJournals(); err != nil {
			return imp.result, fmt.Errorf("failed to import journals: %w", err)
		}
	}

	// Import pages
	if !imp.opts.JournalsOnly {
		if err := imp.importPages(); err != nil {
			return imp.result, fmt.Errorf("failed to import pages: %w", err)
		}
	}

	return imp.result, nil
}

// validateLogseqPath checks that the Logseq graph path exists and has expected structure.
func (imp *Importer) validateLogseqPath() error {
	info, err := os.Stat(imp.opts.LogseqPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("path is not a directory")
	}

	// Check for journals or pages directory
	journalsPath := filepath.Join(imp.opts.LogseqPath, "journals")
	pagesPath := filepath.Join(imp.opts.LogseqPath, "pages")

	hasJournals := dirExists(journalsPath)
	hasPages := dirExists(pagesPath)

	if !hasJournals && !hasPages {
		return errors.New("no 'journals' or 'pages' directory found - is this a Logseq graph?")
	}

	return nil
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// importJournals imports all journal files from the journals/ directory.
func (imp *Importer) importJournals() error {
	journalsPath := filepath.Join(imp.opts.LogseqPath, "journals")

	if !dirExists(journalsPath) {
		if imp.opts.Verbose {
			fmt.Println("No journals directory found, skipping journals import")
		}
		return nil
	}

	files, err := os.ReadDir(journalsPath)
	if err != nil {
		return err
	}

	if imp.opts.Verbose {
		fmt.Printf("Found %d files in journals/\n", len(files))
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".md") {
			continue
		}

		filePath := filepath.Join(journalsPath, file.Name())
		if err := imp.importFile(filePath, true); err != nil {
			imp.result.Errors = append(imp.result.Errors, fmt.Errorf("journal %s: %w", file.Name(), err))
			if imp.opts.Verbose {
				fmt.Printf("  Error importing %s: %v\n", file.Name(), err)
			}
		}
	}

	return nil
}

// importPages imports all page files from the pages/ directory.
func (imp *Importer) importPages() error {
	pagesPath := filepath.Join(imp.opts.LogseqPath, "pages")

	if !dirExists(pagesPath) {
		if imp.opts.Verbose {
			fmt.Println("No pages directory found, skipping pages import")
		}
		return nil
	}

	files, err := os.ReadDir(pagesPath)
	if err != nil {
		return err
	}

	if imp.opts.Verbose {
		fmt.Printf("Found %d files in pages/\n", len(files))
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".md") {
			continue
		}

		filePath := filepath.Join(pagesPath, file.Name())
		if err := imp.importFile(filePath, false); err != nil {
			imp.result.Errors = append(imp.result.Errors, fmt.Errorf("page %s: %w", file.Name(), err))
			if imp.opts.Verbose {
				fmt.Printf("  Error importing %s: %v\n", file.Name(), err)
			}
		}
	}

	return nil
}

// importFile imports a single Logseq file.
func (imp *Importer) importFile(filePath string, isJournal bool) error {
	// Parse the file
	doc, err := ParseFile(filePath, isJournal)
	if err != nil {
		return err
	}

	originalTitle := doc.Title

	// Check for duplicates and find unique title
	if imp.store != nil {
		doc.Title = imp.findUniqueTitle(doc.Title)
		if doc.Title != originalTitle {
			imp.result.Renamed = append(imp.result.Renamed, RenameInfo{
				OriginalTitle: originalTitle,
				NewTitle:      doc.Title,
			})
		}
	}

	if imp.opts.Verbose {
		if doc.Title != originalTitle {
			fmt.Printf("  Importing: %s -> %s\n", originalTitle, doc.Title)
		} else {
			fmt.Printf("  Importing: %s\n", doc.Title)
		}
	}

	// Save to database (unless dry run)
	if !imp.opts.DryRun && imp.store != nil {
		if err := imp.store.Save(doc); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}
	}

	// Update counts
	if isJournal {
		imp.result.JournalsImported++
	} else {
		imp.result.PagesImported++
	}

	return nil
}

// findUniqueTitle returns a unique title by appending (2), (3), etc. if needed.
func (imp *Importer) findUniqueTitle(baseTitle string) string {
	title := baseTitle
	suffix := 2

	for {
		_, err := imp.store.LoadDocumentByTitle(title)
		if errors.Is(err, db.ErrDocumentNotFound) {
			// Title is available
			return title
		}
		if err != nil {
			// Some other error - just use the title and let save handle it
			return title
		}

		// Title exists, try with suffix
		title = fmt.Sprintf("%s (%d)", baseTitle, suffix)
		suffix++

		// Safety limit to prevent infinite loop
		if suffix > 1000 {
			return title
		}
	}
}

// Import is a convenience function that creates an Importer and runs the import.
func Import(opts ImportOptions) (*ImportResult, error) {
	importer := NewImporter(opts)
	return importer.Import()
}
