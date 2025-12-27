<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { EditorView, keymap } from '@codemirror/view';
    import { Block } from './block';
    import { createEventDispatcher } from "svelte";

    const dispatch = createEventDispatcher();
    export let block: Block;

    let editorContainer: HTMLDivElement;
    let view: EditorView;

    export function focus() {
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
    }

    onMount(() => {
        view = new EditorView({
            doc: block.content,
            parent: editorContainer,
            extensions: [
                keymap.of([
                    {key: "Tab", run: () => { dispatch('tab', {id: block.id}); return true; } },
                    {key: "Shift-Tab", run: () => { dispatch('shiftTab', {id: block.id}); return true; } },
                    {key: "Enter", run: () => { dispatch('enter', {id: block.id}); return true; } }
                ])
            ]
        });
    });

    onDestroy(() => {
        if (view) {
            view.destroy();
        }
    });
</script>

<main style="--indent-level: {block.indent}">
    <div>· ({block.indent})</div>
    <div bind:this={editorContainer}>
    </div>
</main>

<style>
    main {
        display: flex;
        align-items: center;
        margin-left: calc(var(--indent-level) * 20px);
        margin-top: 4px;
        margin-bottom: 4px;
    }
    main > div:first-child {
        width: 20px;
        text-align: center;
        user-select: none;
        color: #888;
    }
    main > div:last-child {
        flex-grow: 1;
    }
</style>