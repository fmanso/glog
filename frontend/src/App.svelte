<script lang="ts">
    import Router from 'svelte-spa-router';
    import { onMount } from 'svelte';
    import Document from './Document.svelte';
    import OpenDocument from './OpenDocument.svelte';
    import NewDocument from "./NewDocument.svelte";

    onMount(async () => {
    });

    const routes = {
        '/': Document,
        '/open': OpenDocument,
        '/doc/:id': Document,
        '/new': NewDocument,
        '*': Document,
    }
</script>

<main>
    <nav>
        <a href="/">Today's Document</a> |
        <a href="#/open">Open Document</a> |
        <a href="#/new">New Document</a>
    </nav>
    <Router {routes} />
</main>