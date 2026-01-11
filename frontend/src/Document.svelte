<script lang="ts">
    import {onMount} from 'svelte';
    import DocumentUIElement from './components/DocumentUIElement.svelte';
    import Skeleton from './components/Skeleton.svelte';
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
        {#if document && !document.is_journal}
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
            <div class="content-fade-in">
                <DocumentUIElement document={document}></DocumentUIElement>
            </div>
        {:else}
            <Skeleton lines={4} showTitle={false} />
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
        border: 1px solid var(--danger);
        color: var(--danger);
        padding: 10px 16px;
        border-radius: 8px;
        cursor: pointer;
        font-size: 14px;
        font-weight: 500;
        transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease;
    }

    .delete-btn:hover {
        background: var(--danger);
        border-color: var(--danger);
        color: #fff;
    }

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
        align-items: center;
        justify-content: center;
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

    .modal {
        background: var(--surface-2);
        border: 1px solid var(--border);
        padding: 24px;
        border-radius: var(--radius);
        max-width: 400px;
        width: 90%;
        box-shadow: var(--shadow);
        animation: slideUp 0.2s ease;
    }

    .modal h2 {
        margin: 0 0 12px 0;
        color: var(--text);
        font-size: 18px;
        font-weight: 600;
    }

    .modal p {
        margin: 8px 0;
        color: var(--text-dim);
        font-size: 15px;
        line-height: 1.5;
    }

    .modal .warning {
        color: var(--danger);
        font-size: 13px;
        margin-top: 12px;
    }

    .modal-actions {
        display: flex;
        gap: 10px;
        justify-content: flex-end;
        margin-top: 20px;
    }

    .cancel-btn {
        background: var(--surface-3);
        border: 1px solid var(--border);
        color: var(--text);
        padding: 10px 16px;
        border-radius: 8px;
        cursor: pointer;
        font-size: 14px;
        font-weight: 500;
        transition: background 0.15s ease, border-color 0.15s ease;
    }

    .cancel-btn:hover:not(:disabled) {
        background: var(--surface-1);
        border-color: var(--border-strong);
    }

    .confirm-delete-btn {
        background: var(--danger);
        border: 1px solid var(--danger);
        color: #fff;
        padding: 10px 16px;
        border-radius: 8px;
        cursor: pointer;
        font-size: 14px;
        font-weight: 500;
        transition: background 0.15s ease, filter 0.15s ease;
    }

    .confirm-delete-btn:hover:not(:disabled) {
        filter: brightness(1.1);
    }

    .cancel-btn:disabled,
    .confirm-delete-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .content-fade-in {
        animation: contentFadeIn 0.25s ease;
    }

    @keyframes contentFadeIn {
        from { 
            opacity: 0;
            transform: translateY(4px);
        }
        to { 
            opacity: 1;
            transform: translateY(0);
        }
    }
</style>