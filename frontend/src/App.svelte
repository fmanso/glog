<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { LoadTodayDocument } from "../wailsjs/go/main/App";
  import { SaveDocument } from "../wailsjs/go/main/App";
  import type { main } from "../wailsjs/go/models";

  type ParagraphRow = {
    id: string;
    text: string;
    indent: number;
  };

  let paragraphs: ParagraphRow[] = [];
  let refs: Array<HTMLTextAreaElement | null> = [];
  let nextId = 1;
  let docTitle = '';
  let docDate = '';
  let docId = '';
  const SAVE_DELAY = 800;
  let saveTimer: ReturnType<typeof setTimeout> | null = null;

  const getSubtreeEnd = (start: number) => {
    const base = paragraphs[start]?.indent ?? 0;
    let end = start + 1;
    while (end < paragraphs.length && paragraphs[end].indent > base) end++;
    return end;
  };

  const hasChildren = (index: number) => {
    const currentIndent = paragraphs[index]?.indent ?? 0;
    const next = paragraphs[index + 1];
    return next ? next.indent > currentIndent : false;
  };

  const deleteParagraph = async (index: number) => {
    if (paragraphs.length === 0) return;
    if (hasChildren(index)) return;
    const end = getSubtreeEnd(index);
    const fallback = Math.max(0, index - 1);
    paragraphs = [...paragraphs.slice(0, index), ...paragraphs.slice(end)];
    if (paragraphs.length === 0) {
      paragraphs = [{ id: `local-${nextId++}`, text: '', indent: 0 }];
    }
    await setFocus(Math.min(fallback, paragraphs.length - 1));
    scheduleSave();
  };

  const flattenParagraphs = (items: main.ParagraphDto[] = [], depth = 0): ParagraphRow[] => {
    const list = items ?? [];
    return list.flatMap((p) => [
      { id: p.id, text: p.content, indent: depth },
      ...flattenParagraphs(p.children ?? [], depth + 1),
    ]);
  };

  const buildTree = (rows: ParagraphRow[]): main.ParagraphDto[] => {
    const root = { indent: -1, children: [] as any[] };
    const stack = [root];
    rows.forEach((row) => {
      const node: any = { id: row.id, content: row.text, children: [] as any[] };
      while (stack.length > 0 && row.indent <= stack[stack.length - 1].indent) {
        stack.pop();
      }
      const parent = stack[stack.length - 1];
      parent.children.push(node);
      stack.push({ indent: row.indent, children: node.children });
    });
    return root.children;
  };

  const scheduleSave = () => {
    if (saveTimer) clearTimeout(saveTimer);
    saveTimer = setTimeout(saveDocumentNow, SAVE_DELAY);
  };

  const saveDocumentNow = async () => {
    if (!docId) return;
    const body = buildTree(paragraphs);
    const payload: any = {
      id: docId,
      title: docTitle,
      date: docDate,
      body,
    };
    try {
      await SaveDocument(payload);
    } catch (err) {
      console.error('Failed to save document', err);
    }
  };

  onMount(async () => {
    const doc = await LoadTodayDocument();
    docTitle = doc.title ?? '';
    docDate = doc.date ?? '';
    docId = doc.id ?? '';
    const flat = flattenParagraphs(doc.body ?? [], 0);
    paragraphs = flat;
    nextId = flat.length + 1;
    await tick();
    refs.forEach((ref) => {
      if (ref) autoResize(ref);
    });
  });

  const autoResize = (el: HTMLTextAreaElement) => {
    el.style.height = 'auto';
    el.style.height = `${el.scrollHeight}px`;
  };

  const registerRef = (node: HTMLTextAreaElement, index: number) => {
    refs[index] = node;
    autoResize(node);
    return {
      destroy() {
        refs[index] = null;
      }
    };
  };

  const setFocus = async (index: number, caret?: number) => {
    await tick();
    const ref = refs[index];
    if (ref) {
      ref.focus();
      if (caret !== undefined) {
        ref.setSelectionRange(caret, caret);
      }
    }
  };

  const handleKeyDown = async (event: KeyboardEvent, index: number) => {
    if (event.key === 'Enter' && event.shiftKey) {
      event.preventDefault();
      const target = event.target as HTMLTextAreaElement;
      const start = target.selectionStart;
      const end = target.selectionEnd;
      const current = paragraphs[index];
      const updated = `${current.text.slice(0, start)}\n${current.text.slice(end)}`;
      paragraphs = paragraphs.map((p, i) => (i === index ? { ...p, text: updated } : p));
      await setFocus(index, start + 1);
      await tick();
      autoResize(target);
      scheduleSave();
      return;
    }

    if (event.key === 'Enter') {
      event.preventDefault();
      const insertionPoint = getSubtreeEnd(index);
      const current = paragraphs[index];
      const newParagraph: ParagraphRow = { id: `local-${nextId++}`, text: '', indent: current.indent };
      paragraphs = [...paragraphs.slice(0, insertionPoint), newParagraph, ...paragraphs.slice(insertionPoint)];
      await setFocus(insertionPoint, 0);
      scheduleSave();
      return;
    }

    if (event.key === 'Delete') {
      const target = event.target as HTMLTextAreaElement;
      if (target.value.trim() === '' && !hasChildren(index)) {
        event.preventDefault();
        await deleteParagraph(index);
        return;
      }
    }

    if (event.key === 'Backspace') {
      const target = event.target as HTMLTextAreaElement;
      if (target.value.trim() === '' && !hasChildren(index)) {
        event.preventDefault();
        await deleteParagraph(index);
        return;
      }
    }

    if (event.key === 'Tab') {
      event.preventDefault();
      if (index === 0 && !event.shiftKey) return;
      const prevIndent = index > 0 ? paragraphs[index - 1].indent : 0;
      const baseIndent = paragraphs[index].indent;
      const newIndent = event.shiftKey ? Math.max(0, baseIndent - 1) : Math.min(prevIndent + 1, baseIndent + 1);
      const delta = newIndent - baseIndent;
      if (delta === 0) return;
      const end = getSubtreeEnd(index);
      paragraphs = paragraphs.map((p, i) => {
        if (i < index || i >= end) return p;
        return { ...p, indent: Math.max(0, p.indent + delta) };
      });
      await tick();
      for (let i = index; i < end; i++) {
        const ref = refs[i];
        if (ref) autoResize(ref);
      }
      scheduleSave();
    }
  };

  const handleInput = (event: Event, index: number) => {
    const target = event.target as HTMLTextAreaElement;
    const text = target.value;
    paragraphs = paragraphs.map((p, i) => (i === index ? { ...p, text } : p));
    autoResize(target);
    scheduleSave();
  };
</script>

<main class="page dark">
  <section class="editor">
    <header class="editor-header">
      <h1>{docTitle || 'Today\'s document'}</h1>
      <p>{docDate || ""}</p>
    </header>
    {#each paragraphs as paragraph, i (paragraph.id)}
      <div class="bullet-row" style={`margin-left: ${paragraph.indent * 20}px;`}>
        <span class="bullet-dot">•</span>
        <textarea
          class="bullet-input"
          value={paragraph.text}
          placeholder="Type a thought..."
          on:input={(event) => handleInput(event, i)}
          on:keydown={(event) => handleKeyDown(event, i)}
          use:registerRef={i}
          rows="1"
        ></textarea>
      </div>
    {/each}
  </section>
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: "Inter", system-ui, -apple-system, sans-serif;
    background: #05070d;
    color: #e5e7eb;
  }

  main.page.dark {
    display: flex;
    justify-content: center;
    padding: 48px 16px;
  }

  .editor {
    width: 100%;
    max-width: 820px;
    background: radial-gradient(120% 120% at 20% 10%, rgba(99, 102, 241, 0.12), transparent 45%),
      radial-gradient(140% 120% at 80% 10%, rgba(56, 189, 248, 0.1), transparent 40%),
      #0b0f1a;
    border: 1px solid #161b26;
    border-radius: 18px;
    box-shadow: 0 20px 45px rgba(0, 0, 0, 0.6), inset 0 1px 0 rgba(255, 255, 255, 0.02);
    padding: 28px 30px;
  }

  .editor-header {
    margin-bottom: 14px;
    border-bottom: 1px solid #141926;
    padding-bottom: 10px;
  }

  .editor-header h1 {
    margin: 0 0 4px;
    font-size: 20px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: #f9fafb;
  }

  .editor-header p {
    margin: 0;
    color: #9ca3af;
    font-size: 14px;
  }

  .bullet-row {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 6px 0;
    margin-left: -2px;
  }

  .bullet-dot {
    font-size: 17px;
    color: #a5b4fc;
    width: 18px;
    text-align: center;
    flex-shrink: 0;
    margin-top: 8px;
    filter: drop-shadow(0 2px 4px rgba(99, 102, 241, 0.35));
  }

  .bullet-input {
    flex: 1;
    border: 1px solid rgba(255, 255, 255, 0.04);
    background: rgba(255, 255, 255, 0.03);
    border-radius: 12px;
    padding: 10px 12px;
    font-size: 15px;
    outline: none;
    color: #e5e7eb;
    transition: border-color 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
    resize: none;
    line-height: 1.45;
    min-height: 40px;
    overflow: hidden;
  }

  .bullet-input::placeholder {
    color: #6b7280;
  }

  .bullet-input:hover {
    background: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.06);
  }

  .bullet-input:focus {
    background: rgba(255, 255, 255, 0.07);
    border-color: #6366f1;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.22);
  }
</style>
