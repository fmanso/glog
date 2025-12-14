# Quick Start Guide

## Running glog

### Option 1: Main Menu (Easiest)
```powershell
go run cmd/glog/main.go
```

This launches an interactive menu where you can:
- Create new documents
- Browse documents from the last 10 days
- Quit

### Option 2: Direct Commands

**Create a new document:**
```powershell
go run cmd/create/main.go
```

**Browse existing documents:**
```powershell
go run cmd/browse/main.go
```

## Building Executables

Use the build script to create standalone executables:

```powershell
.\build.ps1
```

This creates:
- `glog.exe` - Main menu
- `glog-create.exe` - Create documents
- `glog-browse.exe` - Browse documents

Then run them:
```powershell
.\glog.exe
.\glog-create.exe
.\glog-browse.exe
```

## Features

### Create Documents
- Enter a title and content
- Content supports multiple lines
- Documents are saved with timestamps
- Full-text search indexing

### Browse Documents
- View all documents from the last 10 days
- Navigate with arrow keys or vim-style (j/k)
- Press Enter to view full document details
- See creation date and document ID
- Press Backspace to return to list

### Document Storage
- Documents stored in BoltDB (`glog.db`)
- Set custom path: `$env:GLOG_DB_PATH="path\to\db.db"`
- Time-based indexing for fast retrieval
- Full-text search capability

## Tips

- Use Tab to navigate between fields when creating documents
- Use j/k (vim-style) or arrow keys for navigation
- Press 'r' in browse mode to refresh the document list
- ESC or Ctrl+C to exit at any time
- Documents are indexed by date and searchable terms

## Example Workflow

1. Start the application:
   ```powershell
   go run cmd/glog/main.go
   ```

2. Select "Create New Document"
3. Enter a title and content
4. Press Tab to navigate to buttons
5. Press Enter on "Save"
6. Back at menu, select "Browse Documents"
7. Use arrow keys to navigate
8. Press Enter to view full document
9. Press Backspace to return to list

Enjoy using glog!

