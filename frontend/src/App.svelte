<script lang="ts">
  import DocumentViewer from './components/DocumentViewer.svelte';
  import {LoadJournalsFromTo} from "../wailsjs/go/main/App";

  let documents = [];

  // arg1 is today, arg2 is 7 days ago
  const now: Date = new Date();
  const sevenDaysAgo: Date = new Date();
  sevenDaysAgo.setDate(now.getDate()-7);
  LoadJournalsFromTo(now.toISOString(), sevenDaysAgo.toISOString()).then((docs) => {
    console.log(docs)
    documents = docs;
  });
</script>

<main class="page dark">
  {#each documents as document}
    <DocumentViewer {document} />
  {/each}
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
