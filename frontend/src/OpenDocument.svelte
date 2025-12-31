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

<main class="page open-documents">
    <header class="page-header">
        <div>
            <p class="eyebrow">Browse</p>
            <h1>Open a document</h1>
        </div>
    </header>

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