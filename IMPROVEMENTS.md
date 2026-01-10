# Top 10 Fixes and Improvements for Glog

This document outlines the most important fixes and improvements identified through a comprehensive code review of the Glog note-taking outliner application.

## 6. Implement Document Deletion

**Priority:** HIGH | **Effort:** 4 hours

- **Location:** Missing from `app.go` and frontend
- **Issue:** No way to delete documents/journals - major feature gap
- **Fix:**
  - Add `DeleteDocument(id)` endpoint in `app.go`
  - Implement cleanup logic for all indexes (time, title, references, scheduled, Bleve)
  - Add UI confirmation dialog and delete button
  - Handle orphaned references gracefully
- **Impact:** Essential feature for any note-taking app

**Required implementation:**

```go
// app.go
func (a *App) DeleteDocument(id string) error {
    docID, err := uuid.Parse(id)
    if err != nil {
        return err
    }
    return a.store.Delete(docID)
}

// db/db.go
func (store *DocumentStore) Delete(id uuid.UUID) error {
    // 1. Load document to get metadata
    // 2. Remove from all indexes:
    //    - documents bucket
    //    - time_index
    //    - title_index
    //    - references_index (both directions)
    //    - scheduled_index
    // 3. Remove from Bleve index
}
```

---

## 7. Add Input Validation and Sanitization

**Priority:** HIGH | **Effort:** 3 hours

- **Location:** All endpoints in `app.go`
- **Issue:** No server-side validation of:
  - Max document/title length
  - Max number of blocks
  - Valid UUID formats
- **Fix:** Add validation middleware with proper error messages
- **Impact:** Security, stability, and data integrity

**Validation rules to implement:**
- Title: max 500 characters
- Block content: max 50,000 characters
- Blocks per document: max 10,000
- UUID format validation on all ID parameters

```go
const (
    MaxTitleLength   = 500
    MaxBlockContent  = 50000
    MaxBlocksPerDoc  = 10000
)

func validateDocument(doc *Document) error {
    if len(doc.Title) > MaxTitleLength {
        return fmt.Errorf("title exceeds maximum length of %d", MaxTitleLength)
    }
    if len(doc.Blocks) > MaxBlocksPerDoc {
        return fmt.Errorf("document exceeds maximum of %d blocks", MaxBlocksPerDoc)
    }
    for _, block := range doc.Blocks {
        if len(block.Content) > MaxBlockContent {
            return fmt.Errorf("block content exceeds maximum length")
        }
    }
    return nil
}
```

---

## 8. Implement Export/Import (Markdown Format)

**Priority:** MEDIUM | **Effort:** 6 hours

- **Location:** New functionality in `app.go` + frontend UI
- **Issue:** No way to export notes or migrate to another system (vendor lock-in)
- **Fix:**
  - Export: Convert documents to Markdown with frontmatter (title, date, scheduled tasks)
  - Import: Parse Markdown files and create documents
  - Batch export all journals as ZIP
- **Impact:** Data portability and user trust

**Export format example:**

```markdown
---
title: My Document
date: 2024-01-15T10:30:00Z
type: document
---

# My Document

- First block content
  - Indented child block
  - Another child
- Back to root level

## Scheduled Tasks
- /scheduled 2024-01-20 Complete this task
```

---

## 9. Add Undo/Redo Functionality

**Priority:** MEDIUM | **Effort:** 8 hours

- **Location:** `BlockUIElement.svelte` + backend snapshot system
- **Issue:** No way to recover from accidental deletions or edits. Current 100ms debounce could lose data on crash.
- **Fix:**
  - Enable CodeMirror 6 history extension
  - Consider backend document snapshots (every N saves or time-based)
  - Add Cmd/Ctrl+Z keyboard shortcuts
- **Impact:** User safety and confidence

**Implementation approach:**
1. CodeMirror already has history extension - ensure it's enabled
2. Add document-level undo stack in Svelte store
3. Consider periodic snapshots in BoltDB for recovery

---

## 10. Optimize Rendering Performance for Large Journals

**Priority:** MEDIUM | **Effort:** 5 hours

- **Location:** `Home.svelte` and `DocumentUIElement.svelte`
- **Issue:**
  - Renders all loaded documents/blocks in DOM (no virtual scrolling)
  - CodeMirror instance per block is memory-heavy
  - Performance degrades with 100+ journal entries
- **Fix:**
  - Implement virtual scrolling (svelte-virtual-list or custom)
  - Lazy load CodeMirror instances (only for focused blocks)
  - Add pagination limit to infinite scroll
- **Impact:** Scalability and responsiveness for long-term users

**Virtual scrolling implementation:**

```svelte
<script>
  import VirtualList from 'svelte-virtual-list';
</script>

<VirtualList items={documents} let:item>
  <DocumentUIElement document={item} />
</VirtualList>
```

---

## Summary

| # | Fix | Priority | Effort | Category |
|---|-----|----------|--------|----------|
| 1 | Remove debug logging | CRITICAL | 5 min | Bug |
| 2 | Delete unused word index | HIGH | 20 min | Tech Debt |
| 3 | Fix timezone handling | HIGH | 30 min | Bug |
| 4 | Bleve error handling | HIGH | 2 hrs | Bug |
| 5 | Fix `Ident` typo | MEDIUM | 15 min | Code Quality |
| 6 | Document deletion | HIGH | 4 hrs | Feature |
| 7 | Input validation | HIGH | 3 hrs | Security |
| 8 | Export/Import | MEDIUM | 6 hrs | Feature |
| 9 | Undo/Redo | MEDIUM | 8 hrs | Feature |
| 10 | Virtual scrolling | MEDIUM | 5 hrs | Performance |

**Total Estimated Effort:** ~28.5 hours

---

## Additional Notes

### Strengths of Current Codebase
- Clean layered architecture with good separation of concerns
- Solid database design with proper indexing
- Modern frontend with accessible components
- Good test coverage for backend (~40%)
- Fast and responsive for small to medium datasets

### Other Issues Noted (Lower Priority)
- No database migration system
- Gob encoding not portable across Go versions
- No frontend tests (Svelte components)
- Missing E2E tests
- No conflict resolution for concurrent edits
- CodeMirror + dependencies create large bundle size
