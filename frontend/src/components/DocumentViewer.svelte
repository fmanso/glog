<script lang="ts">
    import {SetParagraphContent, AddNewParagraph} from '../../wailsjs/go/main/App';
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
        AddNewParagraph(document.id, index+1).then(async (para) => {
            console.log('New paragraph added to backend');
            // Add new paragraph after current
            document.body.splice(index + 1, 0, para);
            document = document;
            await tick();
            inputElements[index+1]?.focus();
        });
    } else if (event.key === 'Backspace') {
        if (paragraph.content === '' && (!paragraph.children || paragraph.children.length == 0)) {
            event.preventDefault();
            let index = document.body.indexOf(paragraph);
            if (index > 0) {
                // Remove current paragraph
                document.body.splice(index, 1);
                document = document;
                await tick();
                inputElements[index - 1]?.focus();
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
    }, 300);
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