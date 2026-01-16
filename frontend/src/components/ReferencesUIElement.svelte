<script lang="ts">
    import { onMount } from 'svelte';
    import { GetReferences, OpenDocument, SaveDocument } from '../../wailsjs/go/main/App';
    import type { main } from '../../wailsjs/go/models'
    import DOMPurify from "dompurify";
    import {marked} from "marked";
    import { replaceLinks } from './replaceLinks';
    import BlockUIElement from './BlockUIElement.svelte';
    
    export let title: string = '';
    let references: main.DocumentReferenceDto[] = [];
    
    // Editing state
    let currentEditingId: string | null = null;
    let editingDocument: main.DocumentDto | null = null;
    let editingBlock: main.BlockDto | null = null;
    let loadingBlockId: string | null = null;

    let lastTitle = '';
    let requestId = 0;
    let mounted = false;

    onMount(() => {
        mounted = true;
        // Trigger load if title was already set before mount.
        if (title && title !== lastTitle) {
            lastTitle = title;
            loadReferences(title);
        }
    });

    // Cache results across component instances (Home renders DocumentUIElement in a list).
    const globalKey = '__glog_references_cache__';
    const globalStore = (globalThis as any)[globalKey] ?? ((globalThis as any)[globalKey] = {
        cache: new Map<string, main.DocumentReferenceDto[]>(),
        inflight: new Map<string, Promise<main.DocumentReferenceDto[]>>()
    });

    const cache: Map<string, main.DocumentReferenceDto[]> = globalStore.cache;
    const inflight: Map<string, Promise<main.DocumentReferenceDto[]>> = globalStore.inflight;

    // Each time title changes, ask for references again (deduped)
    $: if (mounted && title && title !== lastTitle) {
        lastTitle = title;
        loadReferences(title);
    }

    async function loadReferences(nextTitle: string) {
        const thisRequest = ++requestId;

        const cached = cache.get(nextTitle);
        if (cached !== undefined) {
            references = cached;
            return;
        }

        // Check for in-flight request
        let promise = inflight.get(nextTitle);
        if (!promise) {
            // Atomically create and store the promise before any await
            promise = (async () => {
                try {
                    const result = await GetReferences(nextTitle);
                    return result ?? [];
                } catch (err) {
                    console.error(`[ReferencesUIElement] Backend error for: ${nextTitle}`, err);
                    return [];
                }
            })();
            inflight.set(nextTitle, promise);
        }

        try {
            const result = await promise;
            cache.set(nextTitle, result);
            inflight.delete(nextTitle);

            // Ignore stale responses if title changed mid-flight.
            if (thisRequest !== requestId) {
                return;
            }
            references = result;
        } catch (err) {
            console.error(`[ReferencesUIElement] Unexpected error loading references for: ${nextTitle}`, err);
            inflight.delete(nextTitle);
            cache.set(nextTitle, []);
            if (thisRequest === requestId) {
                references = [];
            }
        }
    }

    function requestEdit(id: string | null) {
        const wasEditing = currentEditingId !== null;
        currentEditingId = id;
        if (id === null) {
            editingDocument = null;
            editingBlock = null;
            if (wasEditing) {
                handleEditExit();
            }
        }
    }

    async function startEditing(docId: string, blockId: string) {
        if (loadingBlockId) return;
        
        loadingBlockId = blockId;
        try {
            const doc = await OpenDocument(docId);
            editingDocument = doc;
            // BlockReferenceDto uses Pascal case (Id), BlockDto uses lowercase (id)
            editingBlock = doc.blocks.find((b: main.BlockDto) => b.id === blockId) || null;
            
            if (editingBlock) {
                currentEditingId = blockId;
            }
        } catch (err) {
            console.error('Failed to load document for editing:', err);
        } finally {
            loadingBlockId = null;
        }
    }

    async function handleSave() {
        if (!editingDocument || !editingBlock) return;
        
        try {
            // Update the block in the document
            const blockIndex = editingDocument.blocks.findIndex((b: main.BlockDto) => b.id === editingBlock!.id);
            if (blockIndex !== -1) {
                editingDocument.blocks[blockIndex] = editingBlock;
            }
            
            await SaveDocument(editingDocument);
        } catch (err) {
            console.error('Failed to save:', err);
        }
    }

    async function handleEditExit() {
        // Invalidate cache for current title and reload references
        cache.delete(title);
        await loadReferences(title);
    }

    const renderBlock = (content: string) =>
        DOMPurify.sanitize(
            marked.parse(
                replaceLinks(content ?? ""), { async: false }) as string)
</script>

{#if references?.length}
<main class="references-panel" aria-label="Document references">
    <p class="section-title">References</p>
    {#each references as ref}
        <div class="reference-item">
            <div class="reference-header">
                <a href={"#/doc/" + ref.Id}>{ref.Title}</a>
            </div>
            {#if ref.Blocks?.length}
                <div class="reference-blocks">
                    {#each ref.Blocks as block}
                        {#if currentEditingId === block.Id && editingBlock}
                            <div class="reference-editor" style="margin-left: {block.Indent * 20}px">
                                <BlockUIElement
                                    block={editingBlock}
                                    {currentEditingId}
                                    {requestEdit}
                                    on:save={handleSave}
                                />
                            </div>
                        {:else}
                            <div 
                                class="reference-block"
                                class:loading={loadingBlockId === block.Id}
                                style="margin-left: {block.Indent * 20}px"
                                role="button"
                                tabindex="0"
                                on:click={() => startEditing(ref.Id, block.Id)}
                                on:keydown={(e) => {
                                    if (e.key === 'Enter' || e.key === ' ') {
                                        e.preventDefault();
                                        startEditing(ref.Id, block.Id);
                                    }
                                }}
                            >
                                {#if loadingBlockId === block.Id}
                                    <span class="loading-text">Loading...</span>
                                {:else}
                                    {@html renderBlock(block.Content)}
                                {/if}
                            </div>
                        {/if}
                    {/each}
                </div>
            {/if}
        </div>
    {/each}
</main>
{/if}

<style>
    .references-panel {
        margin-top: 12px;
        padding: 10px 12px;
        border: none;
        border-radius: 10px;
        background: rgba(255, 255, 255, 0.04);
        box-shadow: none;
        animation: panelFadeIn 0.3s ease;
    }

    @keyframes panelFadeIn {
        from {
            opacity: 0;
            transform: translateY(8px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }

    .section-title {
        margin: 0 0 10px 0;
        font-size: 13px;
        letter-spacing: 0.05em;
        text-transform: uppercase;
        color: var(--text-dim);
    }

    .reference-item {
        padding: 10px 0;
    }

    .reference-item + .reference-item {
        border-top: none;
    }

    .reference-header {
        font-weight: 600;
        margin: 0 0 4px 0;
        color: var(--text);
        line-height: 1.25;
    }

    .reference-header a {
        color: var(--accent);
        text-decoration: none;
        transition: color 0.15s ease;
    }

    .reference-header a:hover {
        color: var(--accent-strong);
        text-decoration: underline;
    }

    .reference-blocks {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .reference-block {
        color: var(--text-dim);
        font-size: 0.95rem;
        line-height: 1.4;
        padding: 2px 4px;
        margin-left: -4px;
        transition: color 0.15s ease, background 0.15s ease;
        cursor: text;
        border-radius: 4px;
    }

    .reference-block:hover {
        color: var(--text);
        background: rgba(255, 255, 255, 0.06);
    }

    .reference-block:focus {
        outline: none;
        background: rgba(255, 255, 255, 0.06);
    }

    .reference-block.loading {
        cursor: wait;
        opacity: 0.7;
    }

    .loading-text {
        color: var(--text-dim);
        font-style: italic;
    }

    .reference-editor {
        margin: 0;
        padding: 2px 0;
    }

    .reference-block :global(p) {
        margin: 0;
    }
</style>
