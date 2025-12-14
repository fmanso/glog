s
# glog - Document Management System

A Go-based document management system with a text user interface (TUI) for creating and managing documents.

## Features

- Create documents with title and content
- Store documents in BoltDB
- Interactive TUI using Bubble Tea framework
- Hierarchical paragraph structure with references

## Installation

1. Clone the repository
2. Install dependencies:
```bash
go mod download
```

## Usage

### Main Menu (Recommended)

Run the main menu application to choose between creating or browsing documents:

```bash
go run cmd/glog/main.go
```

Or build and run:

```bash
go build -o glog.exe cmd/glog/main.go
.\glog.exe
```

### Creating a New Document

Run the create TUI application directly:

```bash
go run cmd/create/main.go
```

Or build and run:

```bash
go build -o glog-create.exe cmd/create/main.go
.\glog-create.exe
```

#### Create Mode Controls

- **Tab / Shift+Tab**: Navigate between fields
- **Enter**: 
  - In Title/Content fields: Move to next field
  - On Save button: Save the document
  - On Cancel button: Exit without saving
- **Left/Right**: Switch between Save and Cancel buttons when focused
- **ESC / Ctrl+C**: Exit the application

### Browse Documents (Last 10 Days)

Run the browse TUI application to view and inspect documents:

```bash
go run cmd/browse/main.go
```

Or build and run:

```bash
go build -o glog-browse.exe cmd/browse/main.go
.\glog-browse.exe
```

#### Browse Mode Controls

**List View:**
- **↑/↓ or j/k**: Navigate through documents
- **Enter**: View document details
- **r**: Refresh the list
- **q/ESC**: Quit

**Detail View:**
- **Backspace/Left/h**: Return to list
- **q/ESC**: Quit

### Database Location

By default, the database is stored in `glog.db` in the current directory. You can specify a custom location using the `GLOG_DB_PATH` environment variable:

```bash
# Windows PowerShell
$env:GLOG_DB_PATH="C:\path\to\your\database.db"
go run cmd/create/main.go

# Windows Command Prompt
set GLOG_DB_PATH=C:\path\to\your\database.db
go run cmd/create/main.go
```

## Project Structure

```
glog/
├── domain/          # Domain models and database logic
│   ├── db.go       # DocumentStore implementation
│   ├── paragraph.go # Document and Paragraph structs
│   └── db_test.go  # Tests
gi├── tui/            # Text User Interface
│   ├── tui.go      # Document creation UI
│   ├── browser.go  # Document browsing/viewing UI
│   └── menu.go     # Main menu UI
├── cmd/
│   ├── glog/       # Main menu application
│   │   └── main.go
│   ├── create/     # Create document application
│   │   └── main.go
│   └── browse/     # Browse documents application
│       └── main.go
├── go.mod
├── build.ps1       # Build script
└── README.md
```

## Dependencies

- [BoltDB](https://github.com/boltdb/bolt) - Embedded key-value database
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [UUID](https://github.com/google/uuid) - UUID generation

## License

[Add your license here]

