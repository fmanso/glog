<script lang="ts">
    import {onMount, tick} from 'svelte';
    import DocumentUIElement from './components/DocumentUIElement.svelte';
    import {LoadJournalToday, OpenDocument, OpenDocumentByTitle} from '../wailsjs/go/main/App'
    import { main } from '../wailsjs/go/models'
    let document : main.DocumentDto;

    export let params: { id?: string, title?: string } = {};

    onMount(async () => {
        if (params.id) {
            document = await OpenDocument(params.id); // Replace with LoadDocumentById when implemented
            return;
        }

        if (params.title) {
            document = await OpenDocumentByTitle(params.title); // Implement this function in Go backend
            return;
        }

        document = await LoadJournalToday();
    });
</script>

<main>
    {#if document}
        <DocumentUIElement document={document}></DocumentUIElement>
    {:else}
        <h1>Loading...</h1>
    {/if}
</main>