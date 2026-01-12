package main

import (
	"flag"
	"fmt"
	"os"

	"glog/import/logseq"
)

const usage = `glog-import - Import Logseq journals and pages into glog

Usage:
  glog-import [flags] <logseq-graph-path>

Arguments:
  logseq-graph-path    Path to the Logseq graph directory containing
                       'journals' and/or 'pages' folders

Flags:
  --db <path>          Path to glog database (default: ./glog.db)
  --journals-only      Import only journals, skip pages
  --pages-only         Import only pages, skip journals
  --dry-run            Preview what would be imported without writing
  --verbose            Show detailed progress for each file
  --help               Show this help message

Examples:
  glog-import ~/Documents/my-logseq-graph
  glog-import --db ~/glog.db --verbose ~/logseq
  glog-import --dry-run ~/logseq
  glog-import --journals-only ~/logseq

Note:
  - Flags must be specified before the path argument
  - When importing documents with titles that already exist in glog,
    a suffix will be added (e.g., "My Page" -> "My Page (2)").
`

func main() {
	// Define flags
	dbPath := flag.String("db", "./glog.db", "Path to glog database")
	journalsOnly := flag.Bool("journals-only", false, "Import only journals, skip pages")
	pagesOnly := flag.Bool("pages-only", false, "Import only pages, skip journals")
	dryRun := flag.Bool("dry-run", false, "Preview import without writing")
	verbose := flag.Bool("verbose", false, "Show detailed progress")
	help := flag.Bool("help", false, "Show help message")

	// Custom usage function
	flag.Usage = func() {
		fmt.Print(usage)
	}

	flag.Parse()

	// Show help if requested
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Check for required positional argument
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: missing required argument <logseq-graph-path>")
		fmt.Fprintln(os.Stderr, "")
		flag.Usage()
		os.Exit(1)
	}

	logseqPath := args[0]

	// Validate conflicting flags
	if *journalsOnly && *pagesOnly {
		fmt.Fprintln(os.Stderr, "Error: --journals-only and --pages-only cannot be used together")
		os.Exit(1)
	}

	// Print header
	fmt.Println("glog-import - Logseq to glog importer")
	fmt.Println("")
	fmt.Printf("Logseq graph: %s\n", logseqPath)
	fmt.Printf("Target database: %s\n", *dbPath)
	if *dryRun {
		fmt.Println("Mode: DRY RUN (no changes will be made)")
	}
	fmt.Println("")

	// Create import options
	opts := logseq.ImportOptions{
		LogseqPath:   logseqPath,
		DBPath:       *dbPath,
		JournalsOnly: *journalsOnly,
		PagesOnly:    *pagesOnly,
		DryRun:       *dryRun,
		Verbose:      *verbose,
	}

	// Run import
	result, err := logseq.Import(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
		os.Exit(1)
	}

	// Print results
	printResults(result, *dryRun)

	// Exit with error code if there were errors
	if len(result.Errors) > 0 {
		os.Exit(1)
	}
}

func printResults(result *logseq.ImportResult, dryRun bool) {
	fmt.Println("")
	if dryRun {
		fmt.Println("=== DRY RUN RESULTS ===")
		fmt.Println("The following would be imported:")
	} else {
		fmt.Println("=== IMPORT COMPLETE ===")
	}
	fmt.Println("")

	fmt.Printf("  Journals imported: %d\n", result.JournalsImported)
	fmt.Printf("  Pages imported:    %d\n", result.PagesImported)
	fmt.Printf("  Total:             %d\n", result.JournalsImported+result.PagesImported)

	if len(result.Renamed) > 0 {
		fmt.Println("")
		fmt.Printf("  Renamed (duplicates): %d\n", len(result.Renamed))
		for _, r := range result.Renamed {
			fmt.Printf("    - \"%s\" -> \"%s\"\n", r.OriginalTitle, r.NewTitle)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Println("")
		fmt.Printf("  Errors: %d\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("    - %v\n", e)
		}
	}

	fmt.Println("")
}
