import {EditorView, keymap, highlightSpecialChars, drawSelection, highlightActiveLine, dropCursor,
  rectangularSelection, crosshairCursor} from "@codemirror/view"
import {EditorState} from "@codemirror/state"
import {defaultHighlightStyle, syntaxHighlighting, indentOnInput, bracketMatching,
  foldKeymap} from "@codemirror/language"
import {defaultKeymap, history, historyKeymap} from "@codemirror/commands"
import {searchKeymap, highlightSelectionMatches} from "@codemirror/search"
import {autocompletion, completionKeymap, closeBrackets, closeBracketsKeymap} from "@codemirror/autocomplete"
import {lintKeymap} from "@codemirror/lint"
import {markdown} from "@codemirror/lang-markdown"
import {marked} from "marked"

interface Block {
  id: string
  indent: number
  view: EditorView
  dom: HTMLElement
  editorContainer: HTMLElement
  preview: HTMLElement
}

class BlockManager {
  blocks: Block[] = []
  container: HTMLElement

  constructor(containerId: string) {
    const el = document.getElementById(containerId)
    if (!el) throw new Error(`Container ${containerId} not found`)
    this.container = el
    
    // Initialize with one block
    this.addBlock(null, 0, "Start document")
  }

  createBlockId(): string {
    if (typeof crypto !== "undefined") {
      if (typeof (crypto as any).randomUUID === "function") {
        return (crypto as any).randomUUID()
      }
      if (typeof crypto.getRandomValues === "function") {
        const array = new Uint32Array(1)
        crypto.getRandomValues(array)
        return array[0].toString(36)
      }
    }
    // Fallback: timestamp-based ID avoids Math.random and keeps IDs reasonably unique
    return Date.now().toString(36)
  }

  addBlock(afterBlockId: string | null, indent: number, initialContent: string = "") {
    const id = this.createBlockId()
    const dom = document.createElement("div")
    dom.className = "editor-block"
    dom.style.marginLeft = `${indent * 20}px`
    
    const bullet = document.createElement("div")
    bullet.className = "bullet"
    bullet.textContent = "•"
    dom.appendChild(bullet)
    
    const editorContainer = document.createElement("div")
    editorContainer.className = "editor-container"
    dom.appendChild(editorContainer)

    const preview = document.createElement("div")
    preview.className = "preview"
    preview.style.display = "none"
    preview.style.flexGrow = "1"
    preview.style.paddingLeft = "4px" // Match editor padding if any
    dom.appendChild(preview)

    const extensions = [
      markdown(),
      highlightSpecialChars(),
      history(),
      drawSelection(),
      dropCursor(),
      EditorState.allowMultipleSelections.of(true),
      indentOnInput(),
      syntaxHighlighting(defaultHighlightStyle, {fallback: true}),
      bracketMatching(),
      closeBrackets(),
      autocompletion(),
      rectangularSelection(),
      crosshairCursor(),
      highlightActiveLine(),
      highlightSelectionMatches(),
      EditorView.lineWrapping,
      EditorView.domEventHandlers({
        blur: () => {
          this.renderPreview(id)
        }
      }),
      keymap.of([
        {key: "Tab", run: () => this.indentBlock(id)},
        {key: "Shift-Tab", run: () => this.unindentBlock(id)},
        {key: "Enter", run: () => this.handleEnter(id)},
        {key: "Backspace", run: () => this.handleBackspace(id)},
        {key: "ArrowUp", run: () => this.focusBlockRelative(id, -1)},
        {key: "ArrowDown", run: () => this.focusBlockRelative(id, 1)},
        ...closeBracketsKeymap,
        ...defaultKeymap,
        ...searchKeymap,
        ...historyKeymap,
        ...foldKeymap,
        ...completionKeymap,
        ...lintKeymap
      ])
    ]

    const view = new EditorView({
      doc: initialContent,
      parent: editorContainer,
      extensions: extensions
    })

    const block: Block = { id, indent, view, dom, editorContainer, preview }

    preview.addEventListener('click', () => {
      this.switchToEditMode(id)
    })

    if (afterBlockId === null) {
      this.blocks.push(block)
      this.container.appendChild(dom)
    } else {
      const index = this.blocks.findIndex(b => b.id === afterBlockId)
      if (index === -1) {
        this.blocks.push(block)
        this.container.appendChild(dom)
      } else {
        this.blocks.splice(index + 1, 0, block)
        // Insert in DOM
        const nextBlock = this.blocks[index + 2]
        if (nextBlock) {
          this.container.insertBefore(dom, nextBlock.dom)
        } else {
          this.container.appendChild(dom)
        }
      }
    }
    
    // Initial render if not focused (but newly added blocks usually get focus immediately after)
    // We'll let the caller handle focus.
    
    return block
  }

  renderPreview(blockId: string) {
    const index = this.blocks.findIndex(b => b.id === blockId)
    if (index === -1) return
    
    const block = this.blocks[index]
    const content = block.view.state.doc.toString()
    
    // If empty, keep editor open? Or show empty placeholder?
    // Logseq shows bullet and empty space.
    if (!content.trim()) {
       // Maybe keep editor open if empty? 
       // Or render empty.
    }
    
    // marked.parse returns string | Promise<string>. We assume sync.
    const html = marked.parse(content) as string
    block.preview.innerHTML = html
    
    block.editorContainer.style.display = "none"
    block.preview.style.display = "block"
  }

  switchToEditMode(blockId: string) {
    const index = this.blocks.findIndex(b => b.id === blockId)
    if (index === -1) return
    
    const block = this.blocks[index]
    block.preview.style.display = "none"
    block.editorContainer.style.display = "block"
    block.view.focus()
  }

  handleEnter(blockId: string): boolean {
    const index = this.blocks.findIndex(b => b.id === blockId)
    if (index === -1) return false
    
    const currentBlock = this.blocks[index]
    const newBlock = this.addBlock(blockId, currentBlock.indent)
    this.switchToEditMode(newBlock.id)
    return true
  }

  handleBackspace(blockId: string): boolean {
    const index = this.blocks.findIndex(b => b.id === blockId)
    if (index === -1) return false
    
    const block = this.blocks[index]
    // Only handle if cursor is at start of document
    const state = block.view.state
    const selection = state.selection.main
    if (selection.from > 0 || selection.to > 0) return false

    if (state.doc.length === 0) {
      // Remove block
      if (this.blocks.length > 1) {
        this.removeBlock(blockId)
        // Focus previous block
        const prevIndex = index - 1
        if (prevIndex >= 0) {
          const prevBlock = this.blocks[prevIndex]
          this.switchToEditMode(prevBlock.id)
          // Set cursor to end
          prevBlock.view.dispatch({selection: {anchor: prevBlock.view.state.doc.length}})
        }
      }
      return true
    } else {
      // Merge with previous block?
      // For now, just focus previous block if at start?
      // Let's stick to removing empty blocks for now.
      // If not empty but at start, maybe merge?
      if (index > 0) {
         const prevBlock = this.blocks[index - 1]
         const currentText = state.doc.toString()
         const prevTextLength = prevBlock.view.state.doc.length
         
         // Append text to previous block
         prevBlock.view.dispatch({
           changes: {from: prevTextLength, insert: currentText},
           selection: {anchor: prevTextLength}
         })
         
         this.removeBlock(blockId)
         this.switchToEditMode(prevBlock.id)
         return true
      }
    }
    return false
  }

  removeBlock(blockId: string) {
    const index = this.blocks.findIndex(b => b.id === blockId)
    if (index === -1) return
    
    const block = this.blocks[index]
    block.view.destroy()
    block.dom.remove()
    this.blocks.splice(index, 1)
  }

  indentBlock(blockId: string): boolean {
    const index = this.blocks.findIndex(b => b.id === blockId)
    if (index === -1) return true 
    
    const block = this.blocks[index]
    
    if (index === 0) return true // Cannot indent first block
    
    const prevBlock = this.blocks[index - 1]
    if (block.indent > prevBlock.indent) return true // Already max indented relative to parent
    
    const currentIndent = block.indent
    
    // Identify subtree
    const subtreeIndices = [index]
    for (let i = index + 1; i < this.blocks.length; i++) {
      if (this.blocks[i].indent > currentIndent) {
        subtreeIndices.push(i)
      } else {
        break
      }
    }
    
    subtreeIndices.forEach(i => {
      this.blocks[i].indent += 1
      this.blocks[i].dom.style.marginLeft = `${this.blocks[i].indent * 20}px`
    })
    
    return true
  }

  unindentBlock(blockId: string): boolean {
    const index = this.blocks.findIndex(b => b.id === blockId)
    if (index === -1) return true
    
    const block = this.blocks[index]
    if (block.indent === 0) return true
    
    const currentIndent = block.indent
    
    // Identify subtree
    const subtreeIndices = [index]
    for (let i = index + 1; i < this.blocks.length; i++) {
      if (this.blocks[i].indent > currentIndent) {
        subtreeIndices.push(i)
      } else {
        break
      }
    }
    
    subtreeIndices.forEach(i => {
      this.blocks[i].indent -= 1
      this.blocks[i].dom.style.marginLeft = `${this.blocks[i].indent * 20}px`
    })
    
    return true
  }

  focusBlockRelative(blockId: string, delta: number): boolean {
    const index = this.blocks.findIndex(b => b.id === blockId)
    if (index === -1) return false
    
    const targetIndex = index + delta
    if (targetIndex >= 0 && targetIndex < this.blocks.length) {
      this.switchToEditMode(this.blocks[targetIndex].id)
      return true
    }
    return false
  }
}

new BlockManager("app")