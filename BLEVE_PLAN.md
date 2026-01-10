# Bleve Integration Plan (Bolt stays core)

## Goal
- Replace Bolt word-index search with Bleve search.
- Keep Bolt for documents, journals/time index, references, scheduling.
- Store Bleve index at `<bolt-db-path>.bleve/` (directory).

## Decisions locked in
- Quotes enable phrase search: `"..."` must match.
- Default semantics: **AND** across unquoted terms.
- Fuzzy: **always on**.
- Ignore non-phrase tokens shorter than **2** chars.
- Title boosted relative to content.
- Analyzer: **single language-agnostic**.
- Index directory: `<boltpath>.bleve` (Option A).

## Search behavior
### Parsing
- Extract quoted phrases (can be multiple).
- Extract remaining whitespace tokens, lowercased.
- Drop tokens with length `< 2` (unless inside a phrase).

### Query construction (AND)
- MUST: each phrase as a PhraseQuery over both fields:
  - `title` (boosted)
  - `content`
- MUST: each token as a fuzzy match over both fields:
  - `title` (boosted)
  - `content`

### Fuzziness
- Token length `2–4`: fuzziness = `1`
- Token length `>= 5`: fuzziness = `2`

### Ranking
- Return document IDs ordered by Bleve score.

## Index schema (language-agnostic)
- One Bleve document per note (ID = Bolt UUID string).
- Fields:
  - `title` (text)
  - `content` (text; blocks joined with `\n`)
  - optional `date`
- Use a language-agnostic analyzer (no Spanish/English-specific stemming).

## Consistency strategy (Bolt ↔ Bleve)
- On `DocumentStore.Save`:
  1) Commit Bolt transaction (doc + time + title + references + scheduling).
  2) After commit succeeds, index/update that document in Bleve.
- If Bleve indexing fails:
  - Do not lose the document change.
  - Log the error.
  - Use reindex to recover.

## Reindex support
Add `DocumentStore.ReindexSearch()`:
- Close existing Bleve index.
- Delete `<boltpath>.bleve/`.
- Recreate index.
- Iterate all Bolt docs (`bucketDocs.ForEach`) and index each.

## Implementation steps
1) Add dependency: `github.com/blevesearch/bleve/v2` to `go.mod` and run `go mod tidy`.
2) Add `db/search_bleve.go`:
   - Wrapper type: open/create index; `IndexDoc(*DocDb)`; `Search(query) ([]uuid.UUID, error)`; `Close()`.
   - Build query: quoted phrases + tokens; conjunction; boosted title; fuzzy per token.
3) Modify `db/db.go`:
   - Remove `wordBlockIndex` / `wordTitleIndex` fields and initialization.
   - Add `search *bleveSearch`.
   - In `NewDocumentStore`, open Bleve at `path + ".bleve"`.
   - In `Close()`, close Bleve then Bolt.
4) Update `DocumentStore.Save`:
   - Remove calls to `wordBlockIndex.save` / `wordTitleIndex.save`.
   - After Bolt txn commits, call `store.search.IndexDoc(docDb)`.
5) Replace `DocumentStore.Search`:
   - Query Bleve and return ranked `[]domain.DocumentID`.
6) Add `DocumentStore.ReindexSearch`.
7) Tests:
   - Replace `db/words_test.go` with Bleve tests:
     - content term match
     - title term match
     - phrase match using quotes
     - fuzzy typo match
   - Ensure each test deletes both `./testX.db` and `./testX.db.bleve/` (use `os.RemoveAll` for the directory).
8) Run `go test ./...`.

## Optional preference
- Indexing full block content means `/scheduled ...` and `[[Title]]` are searchable text. Decide later if you want to strip those tokens during indexing.
