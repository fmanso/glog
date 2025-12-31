<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { EditorView, keymap } from '@codemirror/view';
    import type { main } from '../../wailsjs/go/models';
    import { createEventDispatcher } from "svelte";
    import {autocompletion} from "@codemirror/autocomplete";
    import { marked } from 'marked';
    import DOMPurify from 'dompurify';
    import {SearchDocuments} from "../../wailsjs/go/main/App";
    const dispatch = createEventDispatcher();
    export let block: main.BlockDto;

    let editorContainer: HTMLDivElement;
    let view: EditorView;

    export function focus() {
        isEditing = true;
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

    function handleArrowUp() {
        // If caret is at the first line of the block, dispatch
        const state = view.state;
        const selection = state.selection.main;
        const line = state.doc.lineAt(selection.from);
        if (line.from === 0) {
            dispatch('arrowUp', {id: block.id});
            return true;
        }
        return false;
    }

    function handleArrowDown() {
        // If caret is at the last line of the block, dispatch
        const state = view.state;
        const selection = state.selection.main;
        const line = state.doc.lineAt(selection.from);
        const lastLine = state.doc.lineAt(state.doc.length - 1);
        if (line.from === lastLine.from) {
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
                info: r.id
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
                autocompletion({override: [autocomplete]}),
                keymap.of([
                    {key: "Tab", run: () => { dispatch('tab', {id: block.id}); return true; } },
                    {key: "Shift-Tab", run: () => { dispatch('shiftTab', {id: block.id}); return true; } },
                    {key: "Enter", run: () => { dispatch('enter', {id: block.id}); return true; } },
                    {key: "Backspace", run: () => { dispatch('backspace', {id: block.id}); return getCaretPosition() === 0;}},
                    {key: "ArrowUp", run: () => { return handleArrowUp(); } },
                    {key: "ArrowDown", run: () => { return handleArrowDown();} },
                    {key: "Escape", run: () => { isEditing = false; return true; }},
                ]),
                EditorView.updateListener.of((update) => {
                    if (update.docChanged) {
                        block.content = update.state.doc.toString();
                        triggerDebouncedSave();
                    }
                })
            ],
        });
    });

    onDestroy(() => {
        if (view) {
            view.destroy();
        }
    });

    $: if (view && (block.content ?? "") !== view.state.doc.toString()) {
        setDocString(block.content ?? "");
    }

    $: markdownHtml = DOMPurify.sanitize(marked.parse(block.content ?? "", { async: false }) as string);
    let isEditing = true;
    let saveTimeout: number;
    function triggerDebouncedSave() {
        clearTimeout(saveTimeout);
        saveTimeout = setTimeout(() => {
            dispatch('save', {id: block.id});
        }, 100);
    }
</script>

<main style="--indent-level: {block.indent}">
    <div>· ({block.indent})</div>
    <div style="display: {isEditing ? 'block' : 'none'}; width: 100%;">
        <div bind:this={editorContainer}></div>
    </div>

    {#if !isEditing}
        <div class="markdown-preview" on:dblclick={() => isEditing = true} >
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