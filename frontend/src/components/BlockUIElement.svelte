<script lang="ts">
    import { onMount, onDestroy, tick } from 'svelte';
    import { EditorView, keymap } from '@codemirror/view';
    import type { main } from '../../wailsjs/go/models';
    import { createEventDispatcher } from "svelte";
    import {autocompletion, completionKeymap, startCompletion} from "@codemirror/autocomplete";
    import { marked } from 'marked';
    import DOMPurify from 'dompurify';
    import {SearchDocuments} from "../../wailsjs/go/main/App";
    const dispatch = createEventDispatcher();
    export let block: main.BlockDto;
    export let currentEditingId: string | null;
    export let requestEdit: (id: string | null) => void;

    let editorContainer: HTMLDivElement;
    let view: EditorView;

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
        let word = context.matchBefore(/\[\[\w*/);
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

    onMount(() => {
        view = new EditorView({
            doc: block.content,
            parent: editorContainer,
            extensions: [
                EditorView.lineWrapping,
                autocompletion({override: [autocomplete]}),
                keymap.of([
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
            if (!view) return;
            const next = e.relatedTarget;
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

    function replaceLinks(body: string): string {
        return body.replace(/\[\[(.*?)\]\]/g, (match, p1) => {
            return `<a href="#/doc-title/${p1}">${p1}</a>`;
        });
    }

    $: isEditing = block.id === currentEditingId;
</script>

<main class="block" style="--indent-level: {block.indent}">
    <div class="bullet">·</div>
    <div class="editor-pane" style="display: {isEditing ? 'block' : 'none'}; width: 100%;">
        <div bind:this={editorContainer}></div>
    </div>

    {#if !isEditing}
        <div class="markdown-preview"
             on:mousedown={handlePreviewMouseDown}
             on:dblclick={handlePreviewDblClick}>
            {@html markdownHtml}
        </div>
    {/if}

</main>

<style>
    :global(:root) {
        --indent-size-px: 20px;
    }
    main {
        display: flex;
        align-items: center;
        margin-left: calc(var(--indent-level) * var(--indent-size-px));
        margin-top: 4px;
        margin-bottom: 4px;
    }
    main > div:first-child {
        width: var(--indent-size-px);
        text-align: center;
        user-select: none;
        color: #888;
    }
    main > div:last-child {
        flex-grow: 1;
    }
</style>