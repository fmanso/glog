<script lang="ts">
  import {onMount, tick} from 'svelte';
  import {LoadJournals, LoadJournalToday} from "../wailsjs/go/main/App";
  import {main} from "../wailsjs/go/models";
  import DocumentUIElement from "./components/DocumentUIElement.svelte";

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
      items = [...items, ...nextItems];
      page += 1;
    }
  } finally {
    loading = false;
  }

  console.log(items);
};

onMount(() => {
  let cancelled = false;

  const setup = async () => {
    items = [await LoadJournalToday()]; // Load today's journal first

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
</script>

<main class="home">
  <h1>Home</h1>
  <section class="list">
    {#if items.length === 0 && !loading}
      <p class="empty">Nothing yet. Scroll will load once data exists.</p>
    {/if}

    {#each items as document (document.id)}
      <header class="page-header">
        <div>
          <h1>{document?.title ?? 'Loading…'}</h1>
        </div>
      </header>

      <section class="card blocks-card">
        {#if document}
          <DocumentUIElement document={document}></DocumentUIElement>
        {:else}
          <div class="empty-state">Loading…</div>
        {/if}
      </section>
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

<style>
</style>
