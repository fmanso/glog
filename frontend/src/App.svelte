<script lang="ts">
    import Router from 'svelte-spa-router';
    import { onMount } from 'svelte';
    import Document from './Document.svelte';
    import OpenDocument from './OpenDocument.svelte';

    onMount(async () => {
    });

    const routes = {
        '/': Document,
        '/open': OpenDocument,
        '*': Document,
    }
</script>

<main>
    <nav>
        <a href="#/">Today's Document</a> |
        <a href="#/open">Open Document</a>
    </nav>
    <Router {routes} />
</main>