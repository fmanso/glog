<script lang="ts">
    import Router from 'svelte-spa-router';
    import { onMount } from 'svelte';
    import Document from './Document.svelte';
    import OpenDocument from './OpenDocument.svelte';
    import NewDocument from "./NewDocument.svelte";
    import Home from "./Home.svelte";

    onMount(async () => {
    });

    const routes = {
        '/': Home,
        '/open': OpenDocument,
        '/doc/:id': Document,
        '/doc-title/:title': Document,
        '/new': NewDocument,
        '*': Home,
    }
</script>

<main class="app-shell">
    <nav class="top-nav">
        <a href="#/">Today's Document</a>
        <a href="#/open">Open Document</a>
        <a href="#/new">New Document</a>
    </nav>
    <section class="view-frame">
        <Router {routes} />
    </section>
</main>