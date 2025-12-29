<script lang="ts">
    import { CreateNewDocument } from "../wailsjs/go/main/App";

    let title: string = "New Document";

    async function create() {
        CreateNewDocument(title).then((dto) => {
            window.location.href = "#/doc/" + dto.id;
        });
    }
</script>

<input type="text" bind:value={title} />
<button on:click={create}>Create</button>