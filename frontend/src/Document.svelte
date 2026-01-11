<script lang="ts">
    import {onMount} from 'svelte';
    import DocumentUIElement from './components/DocumentUIElement.svelte';
    import {LoadJournalToday, OpenDocument, OpenDocumentByTitle, DeleteDocument} from '../wailsjs/go/main/App'
    import type { main } from '../wailsjs/go/models'
    import { push } from 'svelte-spa-router';
    let document : main.DocumentDto;

    export let params: { id?: string, title?: string } = {};

    let showDeleteConfirm = false;
    let isDeleting = false;

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

    async function handleDelete() {
        if (!document?.id) return;
        
        isDeleting = true;
        try {
            await DeleteDocument(document.id);
            showDeleteConfirm = false;
            // Navigate back to home after deletion
            push('/');
        } catch (error) {
            console.error('Failed to delete document:', error);
            alert('Failed to delete document. Please try again.');
        } finally {
            isDeleting = false;
        }
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
        {#if document && !document.isJournal}
            <button 
                class="delete-btn" 
                on:click={() => showDeleteConfirm = true}
                title="Delete document"
            >
                Delete
            </button>
        {/if}
    </header>

    <section class="card blocks-card">
        {#if document}
            <DocumentUIElement document={document}></DocumentUIElement>
        {:else}
            <div class="empty-state">Loading…</div>
        {/if}
    </section>
</main>

{#if showDeleteConfirm}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <div class="modal-overlay" on:click={() => showDeleteConfirm = false} role="dialog" aria-modal="true">
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <div class="modal" on:click|stopPropagation role="document">
            <h2>Delete Document</h2>
            <p>Are you sure you want to delete "{document?.title}"?</p>
            <p class="warning">This action cannot be undone.</p>
            <div class="modal-actions">
                <button 
                    class="cancel-btn" 
                    on:click={() => showDeleteConfirm = false}
                    disabled={isDeleting}
                >
                    Cancel
                </button>
                <button 
                    class="confirm-delete-btn" 
                    on:click={handleDelete}
                    disabled={isDeleting}
                >
                    {isDeleting ? 'Deleting...' : 'Delete'}
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    .page-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .delete-btn {
        background: transparent;
        border: 1px solid var(--color-danger, #dc3545);
        color: var(--color-danger, #dc3545);
        padding: 0.5rem 1rem;
        border-radius: 4px;
        cursor: pointer;
        font-size: 0.875rem;
        transition: all 0.2s ease;
    }

    .delete-btn:hover {
        background: var(--color-danger, #dc3545);
        color: white;
    }

    .modal-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.5);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
    }

    .modal {
        background: var(--color-surface, #1e1e1e);
        padding: 1.5rem;
        border-radius: 8px;
        max-width: 400px;
        width: 90%;
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    }

    .modal h2 {
        margin: 0 0 1rem 0;
        color: var(--color-text, #fff);
    }

    .modal p {
        margin: 0.5rem 0;
        color: var(--color-text-secondary, #aaa);
    }

    .modal .warning {
        color: var(--color-danger, #dc3545);
        font-size: 0.875rem;
    }

    .modal-actions {
        display: flex;
        gap: 0.75rem;
        justify-content: flex-end;
        margin-top: 1.5rem;
    }

    .cancel-btn {
        background: transparent;
        border: 1px solid var(--color-border, #444);
        color: var(--color-text, #fff);
        padding: 0.5rem 1rem;
        border-radius: 4px;
        cursor: pointer;
    }

    .cancel-btn:hover:not(:disabled) {
        background: var(--color-border, #444);
    }

    .confirm-delete-btn {
        background: var(--color-danger, #dc3545);
        border: none;
        color: white;
        padding: 0.5rem 1rem;
        border-radius: 4px;
        cursor: pointer;
    }

    .confirm-delete-btn:hover:not(:disabled) {
        background: #c82333;
    }

    .cancel-btn:disabled,
    .confirm-delete-btn:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }
</style>