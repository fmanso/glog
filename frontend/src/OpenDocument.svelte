<script lang="ts">
    import { onMount, createEventDispatcher } from 'svelte';
    import { GetRecentDocuments, SearchDocuments } from '../wailsjs/go/main/App'
    import type { main } from '../wailsjs/go/models'
    import { push } from 'svelte-spa-router';

    export let isOpen: boolean = true;

    const dispatch = createEventDispatcher();

    let documents: main.DocumentSummaryDto[] = [];
    let searchQuery: string = '';
    let isSearching: boolean = false;
    let showingRecents: boolean = true;
    let searchInputRef: HTMLInputElement;
    let debounceTimer: number | null = null;

    onMount(async () => {
        await loadRecents();
        // Focus search input when modal opens
        if (searchInputRef) {
            searchInputRef.focus();
        }
    });

    async function loadRecents() {
        showingRecents = true;
        let docs = await GetRecentDocuments(10);
        documents = docs || [];
    }

    async function onSearch(query: string) {
        if (!query.trim()) {
            await loadRecents();
            return;
        }

        showingRecents = false;
        isSearching = true;
        try {
            documents = await SearchDocuments(query);
        } finally {
            isSearching = false;
        }
    }

    function handleInput(event: Event) {
        const value = (event.target as HTMLInputElement).value;
        
        // Clear existing debounce timer
        if (debounceTimer) {
            clearTimeout(debounceTimer);
        }
        
        // Debounce search to avoid too many requests
        debounceTimer = setTimeout(() => {
            onSearch(value);
        }, 300) as unknown as number;
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === 'Escape') {
            closeModal();
        }
    }

    function closeModal() {
        isOpen = false;
        dispatch('close');
        // Navigate back to home when closing the modal
        push('/');
    }

    function selectDocument(docId: string) {
        closeModal();
        push(`/doc/${docId}`);
    }

    function handleOverlayClick(event: MouseEvent) {
        // Only close if clicking on the overlay itself, not the modal content
        if (event.target === event.currentTarget) {
            closeModal();
        }
    }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <div class="modal-overlay" on:click={handleOverlayClick} role="dialog" aria-modal="true">
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <div class="modal open-document-modal" on:click|stopPropagation role="document">
            <header class="modal-header">
                <div>
                    <p class="eyebrow">{showingRecents ? 'Recent' : 'Search Results'}</p>
                    <h2>Open a document</h2>
                </div>
                <button class="close-btn" on:click={closeModal} title="Close" aria-label="Close">
                    <svg class="icon" viewBox="0 0 24 24" aria-hidden="true">
                        <path d="M18 6 6 18"/>
                        <path d="M6 6 18 18"/>
                    </svg>
                </button>
            </header>

            <div class="search-container">
                <input 
                    type="text" 
                    class="search-input" 
                    placeholder="Search documents..." 
                    bind:value={searchQuery} 
                    bind:this={searchInputRef}
                    on:input={handleInput}
                />
                {#if isSearching}
                    <div class="search-loading">Searching...</div>
                {/if}
            </div>

            <div class="document-list">
                {#if documents.length > 0}
                    <div class="list">
                        {#each documents as document}
                            <button class="list-item" on:click={() => selectDocument(document.id)}>
                                <div class="list-title">{document.title}</div>
                            </button>
                        {/each}
                    </div>
                {:else}
                    <div class="empty-state">
                        {#if showingRecents}
                            No recent documents.
                        {:else}
                            No documents found.
                        {/if}
                    </div>
                {/if}
            </div>
        </div>
    </div>
{/if}

<style>
    .modal-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.6);
        backdrop-filter: blur(8px);
        -webkit-backdrop-filter: blur(8px);
        display: flex;
        align-items: flex-start;
        justify-content: center;
        padding-top: 10vh;
        z-index: 1000;
        animation: fadeIn 0.15s ease;
    }

    @keyframes fadeIn {
        from { opacity: 0; }
        to { opacity: 1; }
    }

    @keyframes slideUp {
        from { 
            opacity: 0;
            transform: translateY(10px) scale(0.98);
        }
        to { 
            opacity: 1;
            transform: translateY(0) scale(1);
        }
    }

    .open-document-modal {
        background: var(--surface-1);
        border: 1px solid var(--border);
        padding: 0;
        border-radius: var(--radius);
        max-width: 560px;
        width: 90%;
        max-height: 70vh;
        display: flex;
        flex-direction: column;
        box-shadow: var(--shadow);
        animation: slideUp 0.2s ease;
        overflow: hidden;
    }

    .modal-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 20px 24px 16px;
        border-bottom: 1px solid var(--border);
    }

    .modal-header h2 {
        margin: 4px 0 0 0;
        color: var(--text);
        font-size: 20px;
        font-weight: 600;
    }

    .modal-header .eyebrow {
        margin: 0;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        font-size: 11px;
        color: var(--text-dim);
    }

    .close-btn {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 32px;
        height: 32px;
        padding: 0;
        border-radius: 8px;
        color: var(--text-dim);
        border: 1px solid var(--border);
        background: var(--surface-2);
        cursor: pointer;
        transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease;
    }

    .close-btn:hover {
        background: var(--surface-3);
        color: var(--text);
        border-color: var(--border-strong);
    }

    .close-btn .icon {
        width: 16px;
        height: 16px;
        stroke: currentColor;
        fill: none;
        stroke-width: 1.8;
    }

    .close-btn .icon path {
        stroke-linecap: round;
        stroke-linejoin: round;
    }

    .search-container {
        padding: 16px 24px;
        position: relative;
    }

    .search-input {
        width: 100%;
        padding: 12px 14px;
        border-radius: var(--radius);
        border: 1px solid var(--border);
        background: var(--surface-2);
        color: var(--text);
        font-family: inherit;
        font-size: 15px;
        transition: border-color 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
        caret-color: var(--accent);
    }

    .search-input:focus {
        border-color: var(--accent);
        outline: none;
        box-shadow: 0 0 0 3px var(--accent-weak);
        background: var(--surface-1);
    }

    .search-loading {
        position: absolute;
        right: 36px;
        top: 50%;
        transform: translateY(-50%);
        font-size: 12px;
        color: var(--text-dim);
    }

    .document-list {
        flex: 1;
        overflow-y: auto;
        padding: 0 12px 12px;
    }

    .list {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .list-item {
        display: block;
        width: 100%;
        text-align: left;
        padding: 12px 16px;
        background: var(--surface-2);
        border: 1px solid var(--border);
        border-radius: 8px;
        cursor: pointer;
        transition: background 0.15s ease, border-color 0.15s ease, transform 0.1s ease;
        font-family: inherit;
        font-size: inherit;
        color: var(--text);
    }

    .list-item:hover {
        background: var(--surface-3);
        border-color: var(--border-strong);
    }

    .list-item:active {
        transform: translateY(1px);
    }

    .list-item:focus-visible {
        outline: 2px solid var(--accent);
        outline-offset: 2px;
    }

    .list-title {
        font-weight: 500;
    }

    .empty-state {
        text-align: center;
        color: var(--text-dim);
        padding: 24px;
    }
</style>