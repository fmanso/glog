<script lang="ts">
    import { onMount } from 'svelte';
    import { GetReferences } from '../../wailsjs/go/main/App';
    import type { main } from '../../wailsjs/go/models'
    import DOMPurify from "dompurify";
    import {marked} from "marked";
    import { replaceLinks } from './replaceLinks';
    export let title: string = '';
    let references: main.DocumentReferenceDto[] = [];

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

    const renderBlock = (content: string) =>
        DOMPurify.sanitize(
            marked.parse(
                replaceLinks(content ?? ""), { async: false }) as string)
</script>

{#if references?.length}
<main class="references-panel" aria-label="Document references">
        <div class="references-header">
            <h3>References</h3>
        </div>
        {#each references as ref}
            <div class="reference-item">
                <div class="reference-header">
                    <a href={"#/doc/" + ref.Id}>{ref.Title}</a>
                </div>
                {#if ref.Blocks?.length}
                    <div class="reference-blocks">
                        {#each ref.Blocks as block}
                            <div class="reference-block">
                                {@html renderBlock(block.Content)}
                            </div>
                        {/each}
                    </div>
                {:else}
                    <div class="reference-block reference-block--empty">No specific block provided.</div>
                {/if}
            </div>
        {/each}
</main>
{/if}

<style>
    .references-panel {
        margin-top: 24px;
        padding: 16px 18px;
        border: 1px solid var(--border);
        border-radius: var(--radius);
        background: var(--surface-1);
        box-shadow: var(--shadow);
    }

    .references-header {
        display: flex;
        align-items: baseline;
        gap: 10px;
        margin-bottom: 8px;
    }

    .references-header h3 {
        margin: 0;
        color: var(--text);
        font-size: 16px;
        letter-spacing: 0.01em;
    }

    .references-hint {
        margin: 0;
        color: var(--text-dim);
        font-size: 13px;
    }

    .reference-item {
        border: 1px solid var(--border);
        border-radius: 10px;
        padding: 12px 14px;
        margin-top: 10px;
        background: var(--surface-2);
        box-shadow: 0 12px 30px rgba(0, 0, 0, 0.18);
    }

    .reference-header {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 6px;
    }

    .reference-label {
        padding: 4px 8px;
        border-radius: 999px;
        background: var(--accent-weak);
        color: var(--accent-strong);
        font-size: 12px;
        font-weight: 700;
        letter-spacing: 0.04em;
        text-transform: uppercase;
    }

    .reference-header a {
        color: var(--text);
        font-weight: 600;
        text-decoration: none;
    }

    .reference-header a:hover { color: var(--accent); text-decoration: underline; }

    .reference-blocks {
        display: flex;
        flex-direction: column;
        gap: 8px;
        padding-left: 10px;
        border-left: 1px dashed rgba(255, 255, 255, 0.08);
    }

    .reference-block {
        padding: 10px 12px;
        background: rgba(255, 255, 255, 0.02);
        border-radius: 8px;
        border: 1px solid var(--border);
        color: var(--text);
        line-height: 1.45;
        font-size: 0.95rem;
    }

    .reference-block--empty {
        color: var(--text-dim);
        font-style: italic;
        border: none;
        background: transparent;
    }

    .references-empty {
        margin: 0;
        color: var(--text-dim);
        text-align: center;
        padding: 8px 0;
    }
</style>
