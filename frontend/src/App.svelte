<script lang="ts">
    import Router from 'svelte-spa-router';
    import { onMount } from 'svelte';
    import { location, push } from 'svelte-spa-router';
    import Document from './Document.svelte';
    import OpenDocument from './OpenDocument.svelte';
    import NewDocument from "./NewDocument.svelte";
    import Home from "./Home.svelte";
    import { Quit } from "../wailsjs/runtime/runtime";

    let showOpenModal = false;

    onMount(async () => {
    });

    // Check if we should show the open modal based on the route
    $: {
        if ($location === '/open') {
            showOpenModal = true;
        }
    }

    function handleOpenModalClose() {
        showOpenModal = false;
    }

    function openDocumentModal(e: Event) {
        e.preventDefault();
        showOpenModal = true;
    }

    const routes = {
        '/': Home,
        '/open': Home, // Render Home in background when modal is open
        '/doc/:id': Document,
        '/doc-title/:title': Document,
        '/new': NewDocument,
        '*': Home,
    }
</script>

<main class="app-shell">
    <nav class="top-nav">
        <a class="icon-link" href="#/" title="Today" data-tooltip="Today">
            <svg class="icon" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M4 10.5 12 4l8 6.5V20a1 1 0 0 1-1 1h-5v-5h-4v5H5a1 1 0 0 1-1-1z"/>
            </svg>
        </a>
        <a class="icon-link" href="#/open" on:click={openDocumentModal} title="Open" data-tooltip="Open">
            <svg class="icon" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M4 7h5l2 2h9v9a1 1 0 0 1-1 1H4z"/>
                <path d="M4 7V5a1 1 0 0 1 1-1h5l2 2h5"/>
            </svg>
        </a>
        <a class="icon-link" href="#/new" title="New" data-tooltip="New">
            <svg class="icon" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M12 5v14"/>
                <path d="M5 12h14"/>
            </svg>
        </a>
        <div class="drag-spacer" aria-hidden="true"></div>
        <button class="icon-link close-btn" on:click={Quit} title="Close" aria-label="Close" data-tooltip="Close">
            <svg class="icon" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M18 6 6 18"/>
                <path d="M6 6 18 18"/>
            </svg>
        </button>
    </nav>
    <section class="view-frame">
        <Router {routes} />
    </section>
</main>

<!-- Open Document Modal -->
<OpenDocument isOpen={showOpenModal} on:close={handleOpenModalClose} />