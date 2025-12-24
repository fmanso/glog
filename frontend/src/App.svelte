<script lang="ts">
  import DocumentViewer from './components/DocumentViewer.svelte';
  import { LoadJournal, NewDocument } from "../wailsjs/go/main/App";
  import { main } from "../wailsjs/go/models";
  import Dialog from './Dialog.svelte';

  let document: main.DocumentDto = undefined;

  // Side menu state
  let menuOpen = false;
  let toggleBtnEl: HTMLButtonElement | null = null;
  let menuEl: HTMLElement | null = null;

  function openMenu() {
    menuOpen = true;
    // Focus the menu for keyboard users.
    queueMicrotask(() => menuEl?.focus());
  }

  function closeMenu() {
    menuOpen = false;
    // Return focus to the toggle button.
    queueMicrotask(() => toggleBtnEl?.focus());
  }

  function toggleMenu() {
    menuOpen ? closeMenu() : openMenu();
  }

  function onGlobalKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && menuOpen) {
      e.preventDefault();
      closeMenu();
    }
  }

  let date = new Date();
  LoadJournal(date).then((doc) => {
    console.log(doc);
    document = doc;
  });

  let dialog: HTMLDivElement | null = null;
  function createNewPage(name: string) {
    console.log("Creating new page..." + name);
    dialog.close();
    NewDocument(name).then((doc) => {
        console.log("New document created:");
        console.log(doc);
        document = doc;
    })
  }
</script>

<svelte:window on:keydown={onGlobalKeydown} />

<main class="app-shell dark">
  <header class="topbar">
    <button
      class="icon-btn"
      bind:this={toggleBtnEl}
      type="button"
      on:click={toggleMenu}
      aria-controls="side-menu"
      aria-expanded={menuOpen}
      aria-label={menuOpen ? 'Close menu' : 'Open menu'}>
      <span aria-hidden="true">☰</span>
    </button>

    <div class="title">glog</div>
  </header>

  {#if menuOpen}
    <div class="scrim" on:click={closeMenu} aria-hidden="true" />
  {/if}

  <aside
    id="side-menu"
    class="side-menu"
    class:open={menuOpen}
    bind:this={menuEl}
    tabindex="-1"
    aria-label="Side menu">
    <div class="menu-header">
      <div class="menu-title">Menu</div>
      <button class="icon-btn" type="button" on:click={closeMenu} aria-label="Close menu">
        <span aria-hidden="true">×</span>
      </button>
    </div>

    <nav class="menu-items">
      <button class="menu-btn" type="button" on:click={() => dialog.showModal()}>New Page</button>
    </nav>

    <Dialog bind:dialog
    on:create={(e) => createNewPage(e.detail.name)}
    on:cancel={() => dialog.close()}>
    </Dialog>

    <div class="menu-footer">
      <div class="muted">{date.toDateString()}</div>
    </div>
  </aside>

  <section class="content" aria-label="Main content">
    <DocumentViewer {document} />
  </section>
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: "Inter", system-ui, -apple-system, sans-serif;
    background: #05070d;
    color: #e5e7eb;
  }

  :global(*) {
    box-sizing: border-box;
  }

  main.app-shell.dark {
    --bg: #05070d;
    --panel: #0b1020;
    --panel-2: #0f1730;
    --text: #e5e7eb;
    --muted: rgba(229, 231, 235, 0.7);
    --border: rgba(229, 231, 235, 0.08);
    --shadow: 0 14px 60px rgba(0, 0, 0, 0.55);

    min-height: 100vh;
    background: var(--bg);
  }

  .topbar {
    position: sticky;
    top: 0;
    z-index: 20;

    height: 52px;
    display: flex;
    align-items: center;
    gap: 12px;

    padding: 0 12px;
    border-bottom: 1px solid var(--border);
    background: rgba(5, 7, 13, 0.85);
    backdrop-filter: blur(10px);
  }

  .title {
    font-weight: 600;
    letter-spacing: 0.2px;
    color: var(--text);
  }

  .icon-btn {
    height: 36px;
    min-width: 36px;
    padding: 0 10px;

    border-radius: 10px;
    border: 1px solid var(--border);
    background: rgba(255, 255, 255, 0.02);
    color: var(--text);

    display: inline-flex;
    align-items: center;
    justify-content: center;

    cursor: pointer;
    user-select: none;
  }

  .icon-btn:hover {
    background: rgba(255, 255, 255, 0.05);
  }

  .icon-btn:focus-visible {
    outline: 2px solid rgba(147, 197, 253, 0.8);
    outline-offset: 2px;
  }

  .scrim {
    position: fixed;
    inset: 0;
    z-index: 30;
    background: rgba(0, 0, 0, 0.55);
  }

  .side-menu {
    position: fixed;
    top: 0;
    left: 0;
    z-index: 40;

    height: 100vh;
    width: 280px;
    padding: 12px;

    background: linear-gradient(180deg, var(--panel), var(--panel-2));
    border-right: 1px solid var(--border);
    box-shadow: var(--shadow);

    transform: translateX(-104%);
    transition: transform 180ms ease;

    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .side-menu.open {
    transform: translateX(0);
  }

  .menu-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 6px 4px;
  }

  .menu-title {
    font-weight: 600;
  }

  .menu-items {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .menu-btn {
    width: 100%;
    text-align: left;

    padding: 10px 12px;
    border-radius: 12px;

    border: 1px solid var(--border);
    background: rgba(255, 255, 255, 0.03);
    color: var(--text);

    cursor: pointer;
  }

  .menu-btn:hover {
    background: rgba(255, 255, 255, 0.06);
  }

  .menu-btn:focus-visible {
    outline: 2px solid rgba(147, 197, 253, 0.8);
    outline-offset: 2px;
  }

  .menu-footer {
    margin-top: auto;
    padding: 10px 4px 4px;
    border-top: 1px solid var(--border);
  }

  .muted {
    color: var(--muted);
    font-size: 12px;
  }

  .content {
    display: flex;
    justify-content: center;
    padding: 48px 16px;
  }

  @media (prefers-reduced-motion: reduce) {
    .side-menu {
      transition: none;
    }
  }
</style>
