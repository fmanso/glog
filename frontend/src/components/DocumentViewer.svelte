<script lang="ts">
    import {AddNewParagraph, Indent, SetParagraphContent, Outdent, DeleteParagraphAt, GetReferences} from '../../wailsjs/go/main/App';
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
            // Select the text of the current paragraph after the current cursor position
            let inputElement = event.target as HTMLInputElement;
            let textBeforeCursor = inputElement.value.substring(0, inputElement.selectionStart ?? 0);
            let textAfterCursor = inputElement.value.substring(inputElement.selectionStart ?? 0);
            AddNewParagraph(document.id, textAfterCursor, index + 1).then(async (doc) => {
                console.log('New paragraph added to backend');
                document = doc;
                await tick();
                inputElements[index + 1]?.focus();
                // Put cursor at start
                inputElements[index+1]?.setSelectionRange(0, 0);
                SetParagraphContent(paragraph.id, textBeforeCursor).then(() => {
                    console.log('Paragraph content updated after split');
                    inputElement.value = textBeforeCursor;
                });
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
            // If cursor position at 0, delete current paragraph (backend will append its content to previous paragraph)
            let inputElement = event.target as HTMLInputElement;
            if (inputElement.selectionStart === 0) {
                let index = document.body.indexOf(paragraph);
                if (index > 0) {
                    event.preventDefault();

                    // Snapshot where the caret should land in the previous paragraph:
                    // right after its original content (before merge/appending happens).
                    const prevContentLen = document.body[index - 1]?.content?.length ?? 0;

                    DeleteParagraphAt(document.id, index).then(async (doc) => {
                        console.log('Paragraph deleted in backend');
                        document = doc;
                        await tick();

                        const prevInput = inputElements[index - 1];
                        if (prevInput) {
                            prevInput.focus();
                            // Clamp in case backend changed the previous content unexpectedly.
                            const clamped = Math.min(prevContentLen, prevInput.value?.length ?? 0);
                            prevInput.setSelectionRange(clamped, clamped);
                        }
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

    let references: main.DocumentDto[];

    $: if (document) {
        onDocumentChanged()
    }

    function onDocumentChanged() {
        GetReferences(document.id).then((docs) => {
            references = docs;
        });
    }
</script>

<main>
    {#if !document}
        <h1>Loading...</h1>
    {:else}
        <h1>{document.title}</h1>
        <h2>{document.id}</h2>
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
        <h2>References</h2>
        {#if references && references.length > 0}
            <ul>
                {#each references as ref}
                    <li>{ref.title} (ID: {ref.id})</li>
                {/each}
            </ul>
        {:else}
            <p>No references found.</p>
        {/if}
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