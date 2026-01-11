<script lang="ts">
    export let lines: number = 3;
    export let showTitle: boolean = true;
</script>

<div class="skeleton-container">
    {#if showTitle}
        <div class="skeleton-title"></div>
    {/if}
    <div class="skeleton-blocks">
        {#each Array(lines) as _, i}
            <div class="skeleton-block" style="--delay: {i * 0.05}s; width: {85 - (i % 3) * 15}%">
                <div class="skeleton-bullet"></div>
                <div class="skeleton-line"></div>
            </div>
        {/each}
    </div>
</div>

<style>
    .skeleton-container {
        padding: 0;
        animation: fadeIn 0.3s ease;
    }

    @keyframes fadeIn {
        from { opacity: 0; }
        to { opacity: 1; }
    }

    @keyframes shimmer {
        0% {
            background-position: -200% 0;
        }
        100% {
            background-position: 200% 0;
        }
    }

    .skeleton-title {
        height: 28px;
        width: 40%;
        border-radius: 6px;
        margin-bottom: 20px;
        background: linear-gradient(
            90deg,
            var(--surface-2) 0%,
            var(--surface-3) 50%,
            var(--surface-2) 100%
        );
        background-size: 200% 100%;
        animation: shimmer 1.5s ease-in-out infinite;
    }

    .skeleton-blocks {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .skeleton-block {
        display: flex;
        align-items: center;
        gap: 10px;
        animation: fadeIn 0.3s ease;
        animation-delay: var(--delay);
        animation-fill-mode: backwards;
    }

    .skeleton-bullet {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: var(--border);
        flex-shrink: 0;
        margin-left: 6px;
    }

    .skeleton-line {
        height: 18px;
        flex: 1;
        border-radius: 4px;
        background: linear-gradient(
            90deg,
            var(--surface-2) 0%,
            var(--surface-3) 50%,
            var(--surface-2) 100%
        );
        background-size: 200% 100%;
        animation: shimmer 1.5s ease-in-out infinite;
        animation-delay: var(--delay);
    }
</style>
