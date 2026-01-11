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
    <p class="section-title">References</p>
    {#each references as ref}
        <div class="reference-item">
            <div class="reference-header">
                <a href={"#/doc/" + ref.Id}>{ref.Title}</a>
            </div>
            {#if ref.Blocks?.length}
                <div class="reference-blocks">
                    {#each ref.Blocks as block}
                        <div class="reference-block" style="margin-left: {block.Indent * 20}px">
                            {@html renderBlock(block.Content)}
                        </div>
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
        padding: 2px 0;
        transition: color 0.15s ease;
    }

    .reference-block:hover {
        color: var(--text);
    }

    .reference-block :global(p) {
        margin: 0;
    }
</style>
