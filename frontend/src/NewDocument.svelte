<script lang="ts">
    import { OpenDocumentByTitle } from "../wailsjs/go/main/App";

    let title: string = "New Document";

    async function create() {
        OpenDocumentByTitle(title).then((dto) => {
            window.location.href = "#/doc/" + dto.id;
        });
    }
</script>

<main class="page new-document">
    <header class="page-header">
        <div>
            <p class="eyebrow">Create</p>
            <h1>New document</h1>
        </div>
    </header>

    <section class="card form-card">
        <label for="title">Title</label>
        <input id="title" type="text" bind:value={title} placeholder="Enter title" />
        <button class="primary" on:click={create}>Create</button>
    </section>
</main>
