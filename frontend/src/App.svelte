<script lang="ts">
  import DocumentViewer from './components/DocumentViewer.svelte';
  import {LoadJournal} from "../wailsjs/go/main/App";

  let document: DocumentDto = undefined;

  let date = new Date();
  date.setHours(0, 0, 0, 0);
  LoadJournal(date).then((doc) => {
      console.log(doc);
      document = doc;
  });
</script>

<main class="page dark">
    <DocumentViewer {document} />
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: "Inter", system-ui, -apple-system, sans-serif;
    background: #05070d;
    color: #e5e7eb;
  }

  main.page.dark {
    justify-content: center;
    padding: 48px 16px;
  }
</style>
