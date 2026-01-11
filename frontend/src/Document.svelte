<script lang="ts">
    import {onMount} from 'svelte';
    import DocumentUIElement from './components/DocumentUIElement.svelte';
    import {LoadJournalToday, OpenDocument, OpenDocumentByTitle} from '../wailsjs/go/main/App'
    import type { main } from '../wailsjs/go/models'
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

    let paramsKey = '';

    onMount(async () => {
        // Remember initial params so reactive block won't double-load.
        paramsKey = `${params?.id ?? ''}|${params?.title ?? ''}`;
        await loadDocument();
    });

    // React to actual route param changes (avoid double-load on mount)
    $: {
        const nextKey = `${params?.id ?? ''}|${params?.title ?? ''}`;
        if (nextKey !== paramsKey) {
            paramsKey = nextKey;
            loadDocument();
        }
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