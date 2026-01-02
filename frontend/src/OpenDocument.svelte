<script lang="ts">
    import { onMount } from 'svelte';
    import { GetDocumentList, SearchDocuments } from '../wailsjs/go/main/App'
    import { main } from '../wailsjs/go/models'

    let documents : main.DocumentSummaryDto[] = [];

    onMount(async () => {
        console.log("Requesting document list...");
        let docs = await GetDocumentList();
        console.log(docs)
        documents = docs;
    });
    
    let searchQuery: string = '';

    async function onSearch(query: string) {
        documents = await SearchDocuments(query);
    }

    function handleInput(event: Event) {
        const value = (event.target as HTMLInputElement).value;
        onSearch(value);
    }
</script>

<main class="page open-documents">
    <header class="page-header">
        <div>
            <p class="eyebrow">Browse</p>
            <h1>Open a document</h1>
        </div>
    </header>

    <section>
        <input type="text" class="search-input" placeholder="Search documents..." bind:value={searchQuery} on:input={handleInput}/>
    </section>

    <section class="card list-card">
        {#if documents.length > 0}
            <div class="list">
                {#each documents as document}
                    <a class="list-item" href="#/doc/{document.id}">
                        <div class="list-title">{document.title}</div>
                    </a>
                {/each}
            </div>
        {:else}
            <div class="empty-state">No documents found.</div>
        {/if}
    </section>
</main>