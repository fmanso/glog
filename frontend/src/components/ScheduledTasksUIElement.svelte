<script lang="ts">
    import type { main } from '../../wailsjs/go/models';
    import { GetScheduledTasks } from '../../wailsjs/go/main/App';
    import { onMount } from 'svelte';
    let tasks: main.ScheduledTaskDto[] = [];

    onMount(async () => {
        tasks = await GetScheduledTasks();
    });

    function formatDate(dateString: string) {
        const options: Intl.DateTimeFormatOptions = { year: 'numeric', month: 'short', day: 'numeric' };
        return new Date(dateString).toLocaleDateString('en-US', options);
    }

    function cleanDescription(description: string) {
        const withoutSchedule = description
            .replace(/\/?scheduled\s+\d{4}-\d{2}-\d{2}/gi, '')
            .trim();
        return withoutSchedule.replace(/(?:\r\n|\r|\n)/g, '<br>');
    }
</script>

{#if tasks && tasks.length}
    <main class="tasks">
        <section class="scheduled">
            <p class="section-title">Scheduled for the following days</p>

                {#each tasks as task}
                    <div class="task-item">
                        <div class="task-title">
                            <a class="task-link" href={`#/doc/${task.doc_id}`}>{task.title}</a>
                        </div>
                        <div class="task-desc">{@html cleanDescription(task.description)}</div>
                        {#if task.due_date}
                            <div class="task-meta">
                                <span class="pill">Scheduled {formatDate(task.due_date)}</span>
                            </div>
                        {/if}
                    </div>
                {/each}
        </section>
    </main>
{/if}


<style lang="css">
    .scheduled {
        margin-top: 12px;
        padding: 10px 12px;
        border: none;
        border-radius: 10px;
        /* Slightly lighter than app background */
        background: rgba(255, 255, 255, 0.04);
        box-shadow: none;
        animation: panelFadeIn 0.3s ease;
    }

    @keyframes panelFadeIn {
        from {
            opacity: 0;
            transform: translateY(8px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }

    .section-title {
        margin: 0 0 10px 0;
        font-size: 13px;
        letter-spacing: 0.05em;
        text-transform: uppercase;
        color: var(--text-dim);
    }

    .tasks { color: var(--text); }

    .task-item { padding: 10px 0; }

    .task-item + .task-item { border-top: none; }

    .task-title {
        font-weight: 600;
        margin: 0 0 4px 0;
        color: var(--text);
        line-height: 1.25;
    }

    .task-link { 
        color: var(--accent); 
        text-decoration: none;
        transition: color 0.15s ease;
    }
    .task-link:hover { 
        color: var(--accent-strong); 
        text-decoration: underline; 
    }

    .task-desc {
        margin: 0;
        color: var(--text-dim);
        font-size: 0.95rem;
        line-height: 1.4;
        white-space: pre-line;
        padding-left: 0;
        border-left: none;
        transition: color 0.15s ease;
    }

    .task-item:hover .task-desc {
        color: var(--text);
    }

    .task-meta { padding-left: 0; margin-top: 6px; }

    .pill {
        display: inline-block;
        padding: 2px 10px;
        border-radius: 999px;
        background: var(--accent-weak);
        color: var(--accent-strong);
        border: none;
        font-size: 12px;
        font-weight: 600;
        letter-spacing: 0.01em;
        transition: background 0.15s ease, transform 0.1s ease;
    }

    .pill:hover {
        background: var(--accent);
        color: #07111f;
    }

    .empty { margin: 4px 0 0 0; color: var(--text-dim); }
</style>