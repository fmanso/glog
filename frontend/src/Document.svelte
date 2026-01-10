<script lang="ts">
    import {onMount, tick} from 'svelte';
    import DocumentUIElement from './components/DocumentUIElement.svelte';
    import {LoadJournalToday, OpenDocument, OpenDocumentByTitle} from '../wailsjs/go/main/App'
    import { main } from '../wailsjs/go/models'
    let document : main.DocumentDto;

    export let params: { id?: string, title?: string } = {};

    async function loadDocument() {
        if (params.id) {
            document = await OpenDocument(params.id);
            return;
        }

        if (params.title) {
            document = await OpenDocumentByTitle(params.title);
            return;
        }

        document = await LoadJournalToday();
    }

    onMount(async () => {
        await loadDocument();
    });

    // React to params changes (when navigating between documents)
    $: if (params) {
        loadDocument();
    }
</script>

<main class="page document-view">
    <header class="page-header">
        <div>
            <h1>{document?.title ?? 'Loading…'}</h1>
        </div>
    </header>

    <section class="card blocks-card">
        {#if document}
            <DocumentUIElement document={document}></DocumentUIElement>
        {:else}
            <div class="empty-state">Loading…</div>
        {/if}
    </section>
</main>