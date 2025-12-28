<script lang="ts">
    import { onMount } from 'svelte';
    import DocumentUIElement from './components/DocumentUIElement.svelte';
    import { LoadJournalToday } from '../wailsjs/go/main/App'
    import { main } from '../wailsjs/go/models'
    let document : main.DocumentDto;

    onMount(async () => {
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