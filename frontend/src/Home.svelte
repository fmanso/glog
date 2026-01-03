<script lang="ts">
    import {onMount } from 'svelte';
    import DocumentUIElement from './components/DocumentUIElement.svelte';
    import { LoadJournals } from '../wailsjs/go/main/App'
    import { main } from '../wailsjs/go/models'
    let documents : main.DocumentDto[];
    onMount(async () => {
    });
</script>

<main class="page document-view">
    <header class="page-header">
        <div>
            <p class="eyebrow">Document</p>
<!--            <h1>{document?.title ?? 'Loading…'}</h1>-->
        </div>
    </header>

    <section class="card blocks-card">
        <!--{#if document}-->
        <!--    <DocumentUIElement document={document}></DocumentUIElement>-->
        <!--{:else}-->
        <!--    <div class="empty-state">Loading…</div>-->
        <!--{/if}-->
    </section>
</main>