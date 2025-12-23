<script lang="ts">
    import { createEventDispatcher} from "svelte";


    export let dialog: HTMLDialogElement;
    const dispatch = createEventDispatcher<{ create: {name: string}; cancel: void}>();
    let pageName = '';
</script>

<dialog class="newpage-dialog" bind:this={dialog} on:close>
    <form method="dialog" class="newpage-form" on:submit|preventDefault={() => dispatch('create', {name: pageName})}>
        <header class="newpage-header">
            <h2 class="newpage-title">New page</h2>
            <p class="newpage-subtitle">Enter a name for the page.</p>
        </header>

        <section class="newpage-body">
            <label class="field">
                <span class="label">Page name</span>
                <input
                    bind:value={pageName}
                    class="input"
                    name="pageName"
                    type="text"
                    placeholder="Untitled"
                    autocomplete="off"
                    autofocus
                />
            </label>
            <slot />
        </section>

        <footer class="newpage-footer">
            <button class="btn btn-ghost" on:click={() => dispatch('cancel')}>Cancel</button>
            <button class="btn btn-primary" type="submit">Create</button>
        </footer>
    </form>
</dialog>

<style>
    /* Dialog shell */
    dialog.newpage-dialog {
        border: 1px solid rgba(255, 255, 255, 0.12);
        border-radius: 12px;
        padding: 0;
        width: min(520px, calc(100vw - 2rem));
        background: rgba(20, 22, 28, 0.98);
        color: rgba(255, 255, 255, 0.92);
        box-shadow:
            0 24px 60px rgba(0, 0, 0, 0.6),
            0 8px 18px rgba(0, 0, 0, 0.35);
        overflow: hidden;
    }

    dialog.newpage-dialog::backdrop {
        background: rgba(0, 0, 0, 0.55);
        backdrop-filter: blur(2px);
    }

    /* Form layout */
    .newpage-form {
        display: grid;
        grid-template-rows: auto 1fr auto;
        gap: 0;
        min-height: 0;
    }

    .newpage-header {
        padding: 1.1rem 1.25rem 0.75rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
        background: linear-gradient(
            180deg,
            rgba(255, 255, 255, 0.03),
            rgba(255, 255, 255, 0)
        );
    }

    .newpage-title {
        margin: 0;
        font-size: 1.05rem;
        font-weight: 650;
        letter-spacing: 0.2px;
    }

    .newpage-subtitle {
        margin: 0.35rem 0 0;
        font-size: 0.9rem;
        color: rgba(255, 255, 255, 0.65);
        line-height: 1.35;
    }

    .newpage-body {
        padding: 1rem 1.25rem 1.1rem;
    }

    .field {
        display: grid;
        gap: 0.45rem;
    }

    .label {
        font-size: 0.85rem;
        color: rgba(255, 255, 255, 0.72);
    }

    .input {
        width: 100%;
        box-sizing: border-box;
        border-radius: 10px;
        border: 1px solid rgba(255, 255, 255, 0.14);
        background: rgba(255, 255, 255, 0.06);
        color: rgba(255, 255, 255, 0.92);
        padding: 0.7rem 0.8rem;
        outline: none;
        transition: border-color 120ms ease, box-shadow 120ms ease,
            background 120ms ease;
    }

    .input::placeholder {
        color: rgba(255, 255, 255, 0.45);
    }

    .input:focus {
        border-color: rgba(124, 170, 255, 0.85);
        box-shadow: 0 0 0 3px rgba(124, 170, 255, 0.22);
        background: rgba(255, 255, 255, 0.08);
    }

    .newpage-footer {
        display: flex;
        justify-content: flex-end;
        gap: 0.6rem;
        padding: 0.8rem 1.25rem 1rem;
        border-top: 1px solid rgba(255, 255, 255, 0.08);
        background: rgba(255, 255, 255, 0.02);
    }

    .btn {
        border: 1px solid transparent;
        border-radius: 10px;
        padding: 0.55rem 0.85rem;
        font-weight: 600;
        font-size: 0.9rem;
        cursor: pointer;
        user-select: none;
        transition: transform 60ms ease, background 120ms ease,
            border-color 120ms ease, filter 120ms ease;
    }

    .btn:active {
        transform: translateY(1px);
    }

    .btn-ghost {
        background: rgba(255, 255, 255, 0.06);
        border-color: rgba(255, 255, 255, 0.12);
        color: rgba(255, 255, 255, 0.85);
    }

    .btn-ghost:hover {
        background: rgba(255, 255, 255, 0.09);
        filter: brightness(1.03);
    }

    .btn-primary {
        background: linear-gradient(180deg, #3d7eff, #2f63ff);
        color: white;
        border-color: rgba(0, 0, 0, 0.2);
    }

    .btn-primary:hover {
        filter: brightness(1.06);
    }

    @media (prefers-reduced-motion: reduce) {
        .input,
        .btn {
            transition: none;
        }
    }
</style>
