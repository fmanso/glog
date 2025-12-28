<script lang="ts">
    import { tick } from 'svelte';
    import BlockUIElement from './BlockUIElement.svelte';
    import {Block, Document} from './block';
    export let document: Document;
    let blockInstances: Record<string, BlockUIElement> = {};


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
        let newBlock: Block = {
            id: crypto.randomUUID(),
            content: content,
            indent: block.indent
        };

        blockInstances[block.id].removeContentAfterCaret();
        document.blocks.splice(index + 1, 0, newBlock);
        document = document;
        await tick();
        blockInstances[newBlock.id].focus();
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
        await tick();
        blockInstances[prevBlock.id].focus();
        blockInstances[prevBlock.id].setCaretPosition(futureCaretPosition);
        console.log(document.blocks[index - 1]);
    }

    function arrowUpHandler(event: CustomEvent) {
        console.log('arrowUpHandler', event);
        let id = event.detail.id;

        let index = document.blocks.findIndex(b => b.id === id);
        if (index <= 0) {
            return
        }

        let prevBlock = document.blocks[index - 1];
        blockInstances[prevBlock.id].focus();
    }

    function arrowDownHandler(event: CustomEvent) {
        console.log('arrowDownHandler', event);
        let id = event.detail.id;

        let index = document.blocks.findIndex(b => b.id === id);
        if (index >= document.blocks.length - 1) {
            return
        }

        let nextBlock = document.blocks[index + 1];
        blockInstances[nextBlock.id].focus();
    }
</script>

<main>
    <h1>{document.title}</h1>
    {#each document.blocks as blk (blk.id)}
        <BlockUIElement block={blk}
                        bind:this={blockInstances[blk.id]}
                        on:tab={tabHandler}
                        on:shiftTab={shiftTabHandler}
                        on:enter={enterHandler}
                        on:backspace={backspaceHandler}
                        on:arrowUp={arrowUpHandler}
                        on:arrowDown={arrowDownHandler}
        />
    {/each}
</main>
