<script lang="ts">
    import { onMount } from 'svelte';
    import { GetDocumentList } from '../wailsjs/go/main/App'
    import { main } from '../wailsjs/go/models'

    let documents : main.DocumentSummaryDto[] = [];

    onMount(async () => {
        documents = await GetDocumentList();
    });
</script>

<main>
    {#if documents.length > 0}
        {#each documents as document}
            <div>{document.title}</div>
        {/each}
    {:else}
        <h1>No documents found.</h1>
    {/if}
</main>