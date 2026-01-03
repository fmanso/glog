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
            <p class="references-hint">Snippets below come from other documents that mention this one.</p>
        </div>
        {#each references as ref}
            <div class="reference-item">
                <div class="reference-header">
                    <span class="reference-label">Referenced in</span>
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
        margin-top: 2rem;
        padding: 1.25rem;
        border: none;
        border-radius: var(--radius);
        background: var(--panel);
        box-shadow: var(--shadow);
    }

    .references-header {
        display: flex;
        align-items: baseline;
        gap: 0.75rem;
        margin-bottom: 0.5rem;
    }

    .references-header h3 {
        margin: 0;
        color: var(--text);
        font-size: 1.05rem;
        letter-spacing: 0.01em;
    }

    .references-hint {
        margin: 0;
        color: var(--text-dim);
        font-size: 0.9rem;
    }

    .reference-item {
        border: none;
        border-radius: 10px;
        padding: 0.95rem 1rem;
        margin-top: 0.75rem;
        background: var(--panel-2);
        box-shadow: 0 12px 30px rgba(0, 0, 0, 0.18);
    }

    .reference-header {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        margin-bottom: 0.35rem;
    }

    .reference-label {
        padding: 0.25rem 0.5rem;
        border-radius: 999px;
        background: var(--accent-weak);
        color: var(--accent-strong);
        font-size: 0.75rem;
        font-weight: 700;
        letter-spacing: 0.04em;
        text-transform: uppercase;
    }

    .reference-header a {
        color: var(--text);
        font-weight: 600;
        text-decoration: none;
    }

    .reference-header a:hover {
        color: var(--accent);
        text-decoration: underline;
    }

    .reference-blocks {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        padding-left: 0.55rem;
        border-left: 1px dashed rgba(255, 255, 255, 0.08);
    }

    .reference-block {
        padding: 0.6rem 0.75rem;
        background: rgba(255, 255, 255, 0.02);
        border-radius: 8px;
        border: none;
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
        padding: 0.5rem 0;
    }
</style>
