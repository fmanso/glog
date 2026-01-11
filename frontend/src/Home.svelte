<script lang="ts">
  import {onMount, tick} from 'svelte';
  import {LoadJournals, LoadJournalToday} from "../wailsjs/go/main/App";
  import type {main} from "../wailsjs/go/models";
  import DocumentUIElement from "./components/DocumentUIElement.svelte";
  import ScheduledTasksUIElement from "./components/ScheduledTasksUIElement.svelte";

  let items: main.DocumentDto[] = [];
let page = 0;
let hasMore = true;
let loading = false;
let sentinel: HTMLDivElement | null = null;
let observer: IntersectionObserver | null = null;
const PAGE_SIZE = 10;

const fakeFetch = async (pageNum: number): Promise<main.DocumentDto[]> => {
    // Start is the date of today - pageNum * PAGE_SIZE days
    let startDate = new Date();
    startDate.setDate(startDate.getDate() - (pageNum + 1) * PAGE_SIZE);
    // End is startDate + PAGE_SIZE days
    let endDate = new Date(startDate);
    endDate.setDate(endDate.getDate() + PAGE_SIZE - 1);
    // Fetch journals between startDate and endDate
    console.log(`Fetching journals from ${startDate.toISOString()} to ${endDate.toISOString()}`);
    let journals = await LoadJournals(startDate.toISOString(), endDate.toISOString());
    // Sort journals by date descending
    journals.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
    return journals;
};

const loadMore = async () => {
  if (loading || !hasMore) return;
  loading = true;
  try {
    const nextItems = await fakeFetch(page);
    console.log(`Loaded ${nextItems.length} items for page ${page}`);
    if (nextItems.length === 0) {
      hasMore = false;
    } else {
      const existingIds = new Set(items.map((d) => d.id));
      const dedupedNext = nextItems.filter((d) => !existingIds.has(d.id));
      items = [...items, ...dedupedNext];
      page += 1;
    }
  } finally {
    loading = false;
  }

  console.log(items);
};

  let todayDocument: main.DocumentDto | null = null;

  onMount(() => {
  let cancelled = false;

  const setup = async () => {
    // Load today's journal once per mount; keep stable id.
    todayDocument = await LoadJournalToday();
    items = [todayDocument];

    await loadMore();


    observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (entry?.isIntersecting) loadMore();
      },
      { rootMargin: '120px' }
    );

    await tick(); // wait for sentinel binding
    if (!cancelled && sentinel) observer.observe(sentinel);
  };

  setup();

  return () => {
    cancelled = true;
    if (observer && sentinel) observer.unobserve(sentinel);
    observer?.disconnect();
  };
});

function isToday(document: main.DocumentDto): boolean {
    console.log("Checking if document date is today:", document.date);
    const docDate = new Date(document.date);
    const today = new Date();
    const isIt = docDate.getFullYear() === today.getFullYear() &&
           docDate.getMonth() === today.getMonth() &&
           docDate.getDate() === today.getDate();
    console.log(`Document date ${docDate.toDateString()} is today: ${isIt}`);
    return isIt;
}
</script>

<main class="page document-view home">
  <section class="list">
    {#if items.length === 0 && !loading}
      <p class="empty">Nothing yet. Scroll will load once data exists.</p>
    {/if}

    {#each items as document, i (document.id)}
      <header class="page-header">
        <div>
          <h1>{document?.title ?? 'Loading…'}</h1>
        </div>
      </header>

      <section class="card blocks-card">
        {#if document}
          <DocumentUIElement document={document}></DocumentUIElement>
          {#if isToday(document)}
            <ScheduledTasksUIElement></ScheduledTasksUIElement>
          {/if}
        {:else}
          <div class="empty-state">Loading…</div>
        {/if}
      </section>

      {#if i < items.length - 1}
        <hr class="doc-separator" />
      {/if}
    {/each}

    {#if loading}
      <div class="loader">Loading...</div>
    {/if}

    <div class="sentinel" bind:this={sentinel} aria-hidden="true"></div>

    {#if !hasMore && items.length > 0}
      <p class="done">No more items.</p>
    {/if}
  </section>
</main>
