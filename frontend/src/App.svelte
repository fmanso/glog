<script lang="ts">
    import DocumentUIElement from './components/DocumentUIElement.svelte';
    import { Document, Block } from './components/block';

    // Sample document data
    const sampleDocument: Document = {
        id: "doc1",
        title: "Sample Document",
        blocks: [
            { id: "1", content: "This is the first block of text.", indent: 0 },
            { id: "2", content: "This is the second block of text.", indent: 0 },
        ],
    };
</script>

<main>
    <DocumentUIElement document={sampleDocument}></DocumentUIElement>
</main>