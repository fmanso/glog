<script lang="ts">
    import { tick, onMount, onDestroy } from 'svelte';
    import { SaveDocument } from "../../wailsjs/go/main/App";
    import BlockUIElement from './BlockUIElement.svelte';
    import type { main } from '../../wailsjs/go/models';
    export let document: main.DocumentDto;
    let blockInstances: Record<string, BlockUIElement> = {};
    let currentEditingId: string | null = null;
    let containerEl: HTMLElement;

    const setCurrentEditing = (id: string | null) => {
        currentEditingId = id;
    };

    const focusBlock = async (id: string | null) => {
        currentEditingId = id;
        if (!id) return;
        let retries = 3;
        while (retries-- > 0) {
            await tick();
            const inst = blockInstances[id];
            if (inst) {
                await inst.focus();
                return;
            }
        }
    };

    onMount(() => {
        const handleOutsideMouseDown = (e: MouseEvent) => {
            if (!containerEl) return;
            if (!containerEl.contains(e.target as Node)) {
                setCurrentEditing(null);
            }
        };
        window.addEventListener('mousedown', handleOutsideMouseDown, true);
        return () => {
            window.removeEventListener('mousedown', handleOutsideMouseDown, true);
        };
    });

    function shiftTabHandler(event: CustomEvent) {
        console.log('shiftTabHandler', event);
        let id = event.detail.id;

        let index = document.blocks.findIndex(b => b.id === id);
        if (index === 0 || index === -1) {
            return
        }

        let block = document.blocks[index];
        if (block.indent === 0) {
            return;
        }

        block.indent -= 1;
        index += 1;

        for (; index < document.blocks.length; index++) {
            let b = document.blocks[index];
            if (b.indent <= block.indent) {
                break;
            }
            b.indent -= 1;
        }

        document = document;
        console.log(block);
    }

    function tabHandler(event: CustomEvent) {
        console.log('tabHandler', event);
        let id = event.detail.id;

        let index = document.blocks.findIndex(b => b.id === id);
        if (index === 0 || index === -1) {
            return
        }

        let prevBlock = document.blocks[index - 1];
        let block = document.blocks[index];

        if (prevBlock.indent < block.indent) {
            return;
        }

        block.indent += 1;
        index += 1;

        for (; index < document.blocks.length; index++) {
            let b = document.blocks[index];
            if (b.indent <= block.indent - 1) {
                break;
            }
            b.indent += 1;
        }

        document = document;
        console.log(block);
    }

    async function enterHandler(event: CustomEvent) {
        console.log('enterHandler', event);
        let id = event.detail.id;

        let index = document.blocks.findIndex(b => b.id === id);
        if (index === -1) {
            return
        }

        let block = document.blocks[index];
        let content = blockInstances[block.id].getContentAfterCaret();
        // Content should include text after caret position
        let newBlock: main.BlockDto = {
            id: crypto.randomUUID(),
            content: content,
            indent: block.indent
        };

        blockInstances[block.id].removeContentAfterCaret();
        document.blocks.splice(index + 1, 0, newBlock);
        document = document;
        await focusBlock(newBlock.id);
        console.log(newBlock);
    }

    async function backspaceHandler(event: CustomEvent) {
        console.log('backspaceHandler', event);
        let id = event.detail.id;

        let index = document.blocks.findIndex(b => b.id === id);
        if (index === -1) {
            return
        }

        let block = document.blocks[index];

        // If next block is indented cancel handling
        if (index < document.blocks.length - 1) {
            let nextBlock = document.blocks[index + 1];
            if (nextBlock.indent > block.indent) {
                return;
            }
        }


        let caretPosition = blockInstances[block.id].getCaretPosition();
        if (caretPosition !== 0) {
            return;
        }

        if (blockInstances[block.id].getSelectedContent().length > 0) {
            return;
        }

        if (index === 0) {
            return;
        }

        let futureCaretPosition = document.blocks[index - 1].content.length;
        let prevBlock = document.blocks[index - 1];
        prevBlock.content += block.content;
        document.blocks.splice(index, 1);
        document = document;
        await focusBlock(prevBlock.id);
        blockInstances[prevBlock.id].setCaretPosition(futureCaretPosition);
        console.log(document.blocks[index - 1]);
    }

    async function arrowUpHandler(event: CustomEvent) {
        console.log('arrowUpHandler', event);
        let id = event.detail.id;

        let index = document.blocks.findIndex(b => b.id === id);
        if (index <= 0) {
            return
        }

        let prevBlock = document.blocks[index - 1];
        await focusBlock(prevBlock.id);
    }

    async function arrowDownHandler(event: CustomEvent) {
        console.log('arrowDownHandler', event);
        let id = event.detail.id;

        let index = document.blocks.findIndex(b => b.id === id);
        if (index >= document.blocks.length - 1) {
            return
        }

        let nextBlock = document.blocks[index + 1];
        await focusBlock(nextBlock.id);
    }

    async function saveDocument() {
        console.log("Saving document...", document);
        await SaveDocument(document);
    }
</script>

<main bind:this={containerEl}>
    {#each document.blocks as blk (blk.id)}
        <BlockUIElement block={blk}
                        bind:this={blockInstances[blk.id]}
                        currentEditingId={currentEditingId}
                        requestEdit={setCurrentEditing}
                        on:tab={tabHandler}
                        on:shiftTab={shiftTabHandler}
                        on:enter={enterHandler}
                        on:backspace={backspaceHandler}
                        on:arrowUp={arrowUpHandler}
                        on:arrowDown={arrowDownHandler}
                        on:save={saveDocument}
        />
    {/each}
</main>

<style>
    :global(:root) {
        --indent-size-px: 20px;
    }
    .block {
        display: flex;
        align-items: flex-start;
        gap: 10px;
        margin-top: 4px;
        margin-bottom: 4px;
    }
    .bullet {
        width: var(--indent-size-px);
        text-align: center;
        user-select: none;
        color: #888;
    }
    .editor-pane {
        flex-grow: 1;
    }
</style>
