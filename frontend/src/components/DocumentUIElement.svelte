<script lang="ts">
    import { tick, onMount } from 'svelte';
    import { SaveDocument } from "../../wailsjs/go/main/App";
    import BlockUIElement from './BlockUIElement.svelte';
    import type { main } from '../../wailsjs/go/models';
    import ReferencesUIElement from "./ReferencesUIElement.svelte";
    export let document: main.DocumentDto;
    let blockInstances: Record<string, BlockUIElement> = {};
    let currentEditingId: string | null = null;
    let containerEl: HTMLElement;
    
    // Save indicator state
    let saveStatus: 'idle' | 'saving' | 'saved' = 'idle';
    let saveTimeout: ReturnType<typeof setTimeout> | null = null;

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
        // If there is any selected content, do not handle
        if (event.detail.hasSelection) {
            return;
        }

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
        saveStatus = 'saving';
        
        // Clear any existing timeout
        if (saveTimeout) {
            clearTimeout(saveTimeout);
        }
        
        await SaveDocument(document);
        
        saveStatus = 'saved';
        
        // Hide the saved indicator after 2 seconds
        saveTimeout = setTimeout(() => {
            saveStatus = 'idle';
        }, 2000);
    }
</script>

<main class="document-container" bind:this={containerEl}>
    <!-- Save indicator -->
    <div class="save-indicator" class:visible={saveStatus !== 'idle'}>
        {#if saveStatus === 'saving'}
            <span class="save-dot saving"></span>
            <span>Saving</span>
        {:else if saveStatus === 'saved'}
            <svg class="save-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
            <span>Saved</span>
        {/if}
    </div>
    
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

    {#if document }
        <ReferencesUIElement title={document.title}></ReferencesUIElement>
    {/if}
</main>

<style>
    :global(:root) { --indent-size-px: 18px; }
    /* Separación mínima entre bloques: que se sienta como líneas consecutivas */
    .document-container { 
        display: flex; 
        flex-direction: column; 
        gap: 0;
        position: relative;
    }
    
    /* Save indicator */
    .save-indicator {
        position: absolute;
        top: -28px;
        right: 0;
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 12px;
        color: var(--text-dim);
        opacity: 0;
        transform: translateY(4px);
        transition: opacity 0.2s ease, transform 0.2s ease;
        pointer-events: none;
    }
    
    .save-indicator.visible {
        opacity: 1;
        transform: translateY(0);
    }
    
    .save-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: var(--accent);
    }
    
    .save-dot.saving {
        animation: pulse 1s ease-in-out infinite;
    }
    
    @keyframes pulse {
        0%, 100% { opacity: 0.4; }
        50% { opacity: 1; }
    }
    
    .save-check {
        width: 14px;
        height: 14px;
        color: var(--accent);
    }
</style>
