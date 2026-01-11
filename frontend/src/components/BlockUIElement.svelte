<script lang="ts">
    import { onMount, onDestroy, tick } from 'svelte';
    import { replaceLinks } from './replaceLinks';
    import { EditorView, keymap } from '@codemirror/view';
    import type { main } from '../../wailsjs/go/models';
    import { createEventDispatcher } from "svelte";
    import {autocompletion, completionKeymap, startCompletion} from "@codemirror/autocomplete";
    import { marked } from 'marked';
    import DOMPurify from 'dompurify';
    import {SearchDocuments} from "../../wailsjs/go/main/App";
    import flatpickr from "flatpickr";
    import "flatpickr/dist/themes/dark.css";
    import type { Instance } from "flatpickr/dist/types/instance";
    import {history, historyKeymap} from "@codemirror/commands";

    const dispatch = createEventDispatcher();
    export let block: main.BlockDto;
    export let currentEditingId: string | null;
    export let requestEdit: (id: string | null) => void;

    let editorContainer: HTMLDivElement;
    let view: EditorView;
    let pickerAnchor: HTMLDivElement;
    let pickerInstance: Instance | null = null;

    export async function focus() {
        requestEdit(block.id);
        await tick();
        view?.focus();
    }

    export function getContentAfterCaret() {
        const state = view.state;
        const selection = state.selection.main;
        return state.doc.sliceString(selection.to);
    }

    export function removeContentAfterCaret() {
        const state = view.state;
        const selection = state.selection.main;
        const transaction = state.update({
            changes: { from: selection.to, to: state.doc.length, insert: "" }
        });
        view.dispatch(transaction);

        block.content = view.state.doc.toString();
    }

    export function getSelectedContent() {
        const state = view.state;
        const selection = state.selection.main;
        return state.doc.sliceString(selection.from, selection.to);
    }

    export function getCaretPosition() {
        const state = view.state;
        const selection = state.selection.main;
        return selection.from;
    }

    export function setCaretPosition(position: number) {
        const state = view.state;
        const transaction = state.update({
            selection: { anchor: position }
        });
        view.dispatch(transaction);
    }

    export function setDocString(newContent: string) {
        const state = view.state;
        const transaction = state.update({
            changes: { from: 0, to: state.doc.length, insert: newContent }
        });
        view.dispatch(transaction);
    }

    async function startEditingAndFocus() {
        requestEdit(block.id);
        await tick();
        view?.focus();
    }

    function isOnLink(event: MouseEvent) {
        const target = event.target as HTMLElement | null;
        return !!target?.closest('a');
    }

    function handlePreviewMouseDown(event: MouseEvent) {
        if (isOnLink(event)) return;
        event.preventDefault();
        startEditingAndFocus();
    }

    function handlePreviewDblClick(event: MouseEvent) {
        if (isOnLink(event)) return;
        event.preventDefault();
        startEditingAndFocus();
    }

    function handleArrowUp() {
        const state = view.state;
        const selection = state.selection.main;
        const line = state.doc.lineAt(selection.from);
        if (line.from === 0) {
            requestEdit(block.id);
            dispatch('arrowUp', {id: block.id});
            return true;
        }
        return false;
    }

    function handleArrowDown() {
        const state = view.state;
        if (state.doc.length === 0) {
            requestEdit(block.id);
            dispatch('arrowDown', {id: block.id});
            return true;
        }
        const selection = state.selection.main;
        const line = state.doc.lineAt(selection.from);
        const lastLine = state.doc.lineAt(state.doc.length - 1);
        if (line.from === lastLine.from) {
            requestEdit(block.id);
            dispatch('arrowDown', {id: block.id});
            return true;
        }
        return false;
    }

    async function autocomplete(context) {
        // Match `[[` plus any following characters up to a closing bracket so we trigger even with no trailing word
        let word = context.matchBefore(/\[\[[^\]]*/);
        if (!word || (word.from === word.to && !context.explicit)) return null;

        let results = await SearchDocuments(word.text.slice(2));

        let options = results.map(r => (
            {
                label: r.title,
                type: "document",
                apply: `${r.title}]]`,
            }));

        return {
            from: word.from + 2, // Start the completion AFTER the '[['
            options: options,
            // Optional: add a filter so it narrows down as you type
            filter: true
        };
    }

    function checkForScheduledCommand() {
        const state = view.state;
        const selection = state.selection.main;

        if (!selection.empty) {
            destroyPicker();
            return;
        }

        const line = state.doc.lineAt(selection.head);
        const textBefore = line.text.slice(0, selection.head - line.from);

        if (textBefore.endsWith("/scheduled ")) {
            const coords = view.coordsAtPos(selection.head);
            if (coords && editorContainer && pickerAnchor) {
                const rect = editorContainer.getBoundingClientRect();

                pickerAnchor.style.top = (coords.bottom - rect.top) + "px";
                pickerAnchor.style.left = (coords.left - rect.left) + "px";

                if (!pickerInstance) {
                    pickerInstance = flatpickr(pickerAnchor, {
                        defaultDate: new Date(),
                        allowInput: true,
                        locale: {
                            firstDayOfWeek: 1 // Start week on Monday
                        },
                        onChange: (selectedDates, dateStr) => {
                            insertDate(dateStr);
                        }
                    });
                }
                pickerInstance.open();
            }
        } else {
            destroyPicker();
        }
    }

    function destroyPicker() {
        if (pickerInstance) {
            pickerInstance.destroy();
            pickerInstance = null;
        }
    }

    function insertDate(dateStr: string) {
        const state = view.state;
        const selection = state.selection.main;

        const transaction = state.update({
            changes: { from: selection.head, insert: dateStr + " " },
            selection: { anchor: selection.head + dateStr.length + 1 }
        });

        view.dispatch(transaction);
        view.focus();
        destroyPicker();
    }

    const editorTheme = EditorView.theme({
        // Caret and selection colors
        ".cm-content, .cm-content *": { caretColor: "var(--accent)" },
        ".cm-cursor, .cm-dropCursor": { borderLeftColor: "var(--accent)" },
        ".cm-selectionBackground, .cm-content ::selection": { backgroundColor: "var(--accent-weak)" },
        // Reset CodeMirror defaults to match our design
        "&": {
            backgroundColor: "transparent",
            fontSize: "16px",
            lineHeight: "1.6",
        },
        "&.cm-focused": {
            outline: "none",
        },
        ".cm-scroller": {
            fontFamily: "var(--font)",
            lineHeight: "1.6",
            overflow: "visible",
        },
        ".cm-content": {
            padding: "0",
            caretColor: "var(--accent)",
        },
        ".cm-line": {
            padding: "0",
        },
        ".cm-activeLine": {
            backgroundColor: "transparent",
        },
        ".cm-activeLineGutter": {
            backgroundColor: "transparent",
        },
    }, { dark: true });

    onMount(() => {
        view = new EditorView({
            doc: block.content,
            parent: editorContainer,
            extensions: [
                EditorView.lineWrapping,
                editorTheme,
                history(),
                autocompletion({override: [autocomplete]}),
                keymap.of([
                    ...historyKeymap,
                    ...completionKeymap,
                    {key: "Tab", run: () => { dispatch('tab', {id: block.id}); return true; } },
                    {key: "Shift-Tab", run: () => { dispatch('shiftTab', {id: block.id}); return true; } },
                    {key: "Enter", run: () => { dispatch('enter', {id: block.id}); return true; } },
                    {key: "Backspace", run: () => { dispatch('backspace', {id: block.id}); return getCaretPosition() === 0;}},
                    {key: "ArrowUp", run: () => { return handleArrowUp(); } },
                    {key: "ArrowDown", run: () => { return handleArrowDown();} },
                    {key: "Escape", run: () => { requestEdit(null); return true; }},
                ]),
                EditorView.updateListener.of((update) => {
                    if (update.focusChanged && view.hasFocus) {
                        requestEdit(block.id);
                    }
                    if (update.docChanged || update.selectionSet) {
                        checkForScheduledCommand();
                    }
                    if (update.docChanged) {
                        block.content = update.state.doc.toString();
                        triggerDebouncedSave();
                        let inserted = "";
                        update.changes.iterChanges((_, __, ___, ____, insert) => { inserted += insert.toString(); });
                        const pos = view.state.selection.main.from;
                        if (inserted.includes("[") && pos >= 2 && view.state.doc.sliceString(pos - 2, pos) === "[[") {
                            startCompletion(view);
                        }
                    }
                })
            ],
        });

        handleBlur = (e: FocusEvent) => {
            const next = e.relatedTarget as HTMLElement;
            if (next && (next.classList.contains('flatpickr-calendar') || next.closest('.flatpickr-calendar'))) {
                return;
            }

            if (!view) return;
            // const next = e.relatedTarget; // Removed this line as it is redefined above
            if (!(next instanceof Element)) {
                flushSave();
                requestEdit(null);
                return;
            }
            if (view.dom.contains(next)) return;
            if (next.closest('main.block')) return;
            flushSave();
            requestEdit(null);
        };

        view.dom.addEventListener('focusout', handleBlur);
    });

    let handleBlur: (e: FocusEvent) => void;

    onDestroy(() => {
        if (view) {
            if (handleBlur) view.dom.removeEventListener('focusout', handleBlur);
            clearTimeout(saveTimeout);
            view.destroy();
        }
        destroyPicker();
    });

    $: if (view && (block.content ?? "") !== view.state.doc.toString()) {
        setDocString(block.content ?? "");
    }

    $: markdownHtml = DOMPurify.sanitize(
        marked.parse(
            replaceLinks(block.content ?? ""), { async: false }) as string);
    let isEditing = true;
    let saveTimeout: any;
    function triggerDebouncedSave() {
        clearTimeout(saveTimeout);
        saveTimeout = setTimeout(() => {
            dispatch('save', {id: block.id});
        }, 100);
    }

    function flushSave() {
        clearTimeout(saveTimeout);
        dispatch('save', {id: block.id});
    }

    $: isEditing = block.id === currentEditingId;
</script>

<main class={`block ${isEditing ? 'block-editing' : ''} document-block`} style={`--indent-level: ${block.indent}`}>
    <div class="bullet">•</div>
    <div class="editor-pane" style="display: {isEditing ? 'block' : 'none'}; width: 100%; position: relative;">
        <div bind:this={editorContainer}></div>
        <div bind:this={pickerAnchor} style="position: absolute; width: 1px; height: 1px; opacity: 0; pointer-events: none;"></div>
    </div>

    {#if !isEditing}
        <div class="markdown-preview"
             role="textbox"
             tabindex="0"
             aria-label="Edit block"
             on:mousedown={handlePreviewMouseDown}
             on:dblclick={handlePreviewDblClick}
             on:keydown={(e) => {
                 if (e.key === 'Enter' || e.key === ' ') {
                     e.preventDefault();
                     startEditingAndFocus();
                 }
             }}>
            {#if (block.content ?? '').length === 0}
                <span class="empty-placeholder">&nbsp;</span>
            {:else}
                {@html markdownHtml}
            {/if}
        </div>
    {/if}

</main>

<style>
    :global(:root) {
        --indent-size-px: 18px;
    }

    main {
        display: flex;
        align-items: baseline;
        gap: 6px;
        margin-left: calc(var(--indent-level) * var(--indent-size-px));
        margin-top: 0;
        margin-bottom: 0;
        padding: 0;
        border: none;
        background: transparent;
    }

    /* Bullet styling - consistent in both modes */
    .bullet {
        width: 18px;
        min-width: 18px;
        text-align: center;
        user-select: none;
        color: var(--text-dim);
        font-size: 24px;
        line-height: 1.6;
        flex: 0 0 auto;
    }

    /* Content area - both editor and preview */
    .editor-pane,
    .markdown-preview {
        flex: 1;
        min-width: 0;
        font-family: var(--font);
        font-size: 16px;
        line-height: 1.6;
        color: var(--text);
        background: transparent;
        border: none;
        padding: 0;
        margin: 0;
    }

    /* Ensure CodeMirror inherits our styles */
    .editor-pane :global(.cm-editor) {
        background: transparent !important;
        border: none !important;
        padding: 0 !important;
        font-family: var(--font) !important;
        font-size: 16px !important;
        line-height: 1.6 !important;
    }

    .editor-pane :global(.cm-editor.cm-focused) {
        outline: none !important;
        box-shadow: none !important;
        border: none !important;
    }

    .editor-pane :global(.cm-scroller) {
        font-family: var(--font) !important;
        line-height: 1.6 !important;
        overflow: visible !important;
    }

    .editor-pane :global(.cm-content) {
        padding: 0 !important;
        font-family: var(--font) !important;
    }

    .editor-pane :global(.cm-line) {
        padding: 0 !important;
    }

    .editor-pane :global(.cm-activeLine) {
        background-color: transparent !important;
    }

    /* Markdown preview - matches editor exactly */
    .markdown-preview {
        width: 100%;
        min-height: 1.6em;
        cursor: text;
    }

    .markdown-preview:focus {
        outline: none;
    }

    /* Reset markdown element styles for inline feel */
    .markdown-preview :global(p) {
        margin: 0;
        padding: 0;
    }

    .markdown-preview :global(p:last-child) {
        margin-bottom: 0;
    }

    /* Links in preview */
    .markdown-preview :global(a) {
        color: var(--accent);
        text-decoration: none;
    }

    .markdown-preview :global(a:hover) {
        color: var(--accent-strong);
        text-decoration: underline;
    }

    /* Code styling */
    .markdown-preview :global(code) {
        font-family: var(--mono);
        font-size: 0.9em;
        background: var(--surface-2);
        padding: 1px 4px;
        border-radius: 4px;
    }

    .markdown-preview :global(pre) {
        font-family: var(--mono);
        font-size: 0.9em;
        background: var(--surface-2);
        padding: 8px 12px;
        border-radius: 6px;
        margin: 4px 0;
        overflow-x: auto;
    }

    .markdown-preview :global(pre code) {
        background: transparent;
        padding: 0;
    }

    /* Empty state placeholder */
    .empty-placeholder {
        display: inline-block;
        width: 100%;
        min-height: 1.6em;
        color: var(--text-dim);
        opacity: 0.5;
    }

    /* Flatpickr z-index fix */
    :global(.flatpickr-calendar) {
        z-index: 9999 !important;
    }

    /* Block editing state - subtle highlight */
    main.block-editing {
        background: transparent;
        border: none;
    }
</style>
