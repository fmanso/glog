<script lang="ts">
    import { onMount } from 'svelte';
    import { GetDocumentList } from '../wailsjs/go/main/App'
    import { main } from '../wailsjs/go/models'

    let documents : main.DocumentSummaryDto[] = [];

    onMount(async () => {
        console.log("Requesting document list...");
        let docs = await GetDocumentList();
        console.log(docs)
        documents = docs;
    });
</script>

<main>
    {#if documents.length > 0}
        {#each documents as document}
            <div><a href="#/doc/{document.id}">{document.title}</a></div>
        {/each}
    {:else}
        <h1>No documents found.</h1>
    {/if}
</main>