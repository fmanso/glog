<script lang="ts">
    import { GetReferences } from '../../wailsjs/go/main/App';
    import { main } from '../../wailsjs/go/models'
    import DOMPurify from "dompurify";
    import {marked} from "marked";
    import { replaceLinks } from './replaceLinks';
    export let title: string = '';
    let references: main.DocumentReferenceDto[] = [];
    // Each time title changes, ask for references again
    $: if (title) {
        loadReferences(title);
    }

    async function loadReferences(title: string) {
        console.log(`Loading references for ${title}`);
        references = await GetReferences(title);
        console.log(`Found: ${references}`);
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
