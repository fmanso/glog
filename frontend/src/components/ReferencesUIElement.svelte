<script lang="ts">
    import { GetReferences } from '../../wailsjs/go/main/App';
    import { main } from '../../wailsjs/go/models'
    export let title: string = '';
    let references: main.DocumentSummaryDto[] = [];
    // Each time title changes, ask for references again
    $: if (title) {
        loadReferences(title);
    }

    async function loadReferences(title: string) {
        references = await GetReferences(title);
    }
</script>

<main>
    <h3>References</h3>
    {#each references as ref}
        <div class="reference-item">
            <a href={"#/doc/" + ref.id}>{ref.title}</a>
        </div>
    {/each}
</main>