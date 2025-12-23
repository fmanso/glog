<script lang="ts">
    import {AddNewParagraph, Indent, SetParagraphContent, Outdent, DeleteParagraphAt} from '../../wailsjs/go/main/App';
    import {main} from '../../wailsjs/go/models';
    import {tick} from 'svelte';

    export let document: main.DocumentDto;

    let debounceTimer: any;
    let inputElements: HTMLInputElement[] = [];

    function handleClick(paragraph: main.ParagraphDto) {
    }

    async function handleKeyDown(event: KeyboardEvent, paragraph: main.ParagraphDto) {
        if (event.key === 'Enter') {
            event.preventDefault();
            let index = document.body.indexOf(paragraph);
            AddNewParagraph(document.id, index + 1).then(async (doc) => {
                console.log('New paragraph added to backend');
                document = doc;
                await tick();
                inputElements[index + 1]?.focus();
            });
        } else if (event.key === 'Tab') {
            if (!event.shiftKey) {
                // Indent
                event.preventDefault();
                let index = document.body.indexOf(paragraph);
                Indent(document.id, index).then(async (doc) => {
                    console.log('Paragraph indented in backend2');
                    document = doc;
                    await tick();
                    inputElements[index]?.focus();
                });
            } else {
                // Outdent
                event.preventDefault();
                let index = document.body.indexOf(paragraph);
                Outdent(document.id, index).then(async (doc) => {
                    console.log('UnIndent indent in backend2');
                    document = doc;
                    await tick();
                    inputElements[index]?.focus();
                })
            }
        } else if (event.key == 'Backspace') {
            // If cursor position at 0, delete current paragraph
            let inputElement = event.target as HTMLInputElement;
            if (inputElement.selectionStart === 0) {
                let index = document.body.indexOf(paragraph);
                if (index > 0) {
                    event.preventDefault();
                    let index = document.body.indexOf(paragraph);
                    DeleteParagraphAt(document.id, index).then(async (doc) => {
                        console.log('Paragraph deleted in backend');
                        document = doc;
                        await tick();
                        inputElements[index - 1]?.focus();
                    });
                }
            }
        }
    }

    function handleInput(event: Event, paragraph: main.ParagraphDto) {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
            console.log('User stopped typing');
            SetParagraphContent(paragraph.id, paragraph.content).then(() => {
                console.log('Paragraph content saved');
            });
        }, 10);
    }
</script>

<main>
    {#if !document}
        <h1>Loading...</h1>
    {:else}
        <h1>{document.title}</h1>
        <!-- For each paragraph in document.paragraphs, render a <p> element -->
        {#each document.body as paragraph, i}
            <span>{i} - {paragraph.id}</span>
            <input type="text"
                   style="margin-left: {paragraph.indentation * 20}px"
                   bind:this={inputElements[i]}
                   bind:value={paragraph.content}
                   on:click={() => handleClick(paragraph)}
                   on:keydown={(e) => handleKeyDown(e, paragraph)}
                   on:input={(e) => handleInput(e, paragraph)}/>
        {/each}
    {/if}
</main>


<style>
    h1 {
        font-size: 24px;
        margin: 0 0 16px 0;
    }

    input {
        background: transparent;
        border: 1px #ddd solid;
        color: inherit;
        width: 100%;
        outline: none;
        font-size: 1rem;
        font-family: inherit;
        padding: 4px 0;
    }

    main {
        background: #1e293b;
        padding: 24px;
        border-radius: 8px;
        margin-bottom: 24px;
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        justify-content: flex-start;
    }
</style>