<script lang="ts">
    import { onMount } from 'svelte';
    import DocumentUIElement from './components/DocumentUIElement.svelte';
    import {LoadJournalToday, OpenDocument} from '../wailsjs/go/main/App'
    import { main } from '../wailsjs/go/models'
    let document : main.DocumentDto;

    export let params: { id?: string } = {};

    onMount(async () => {
        if (params.id) {
            document = await OpenDocument(params.id); // Replace with LoadDocumentById when implemented
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