<script lang="ts">
    import type { main } from '../../wailsjs/go/models';
    import { GetScheduledTasks, OpenDocument, SaveDocument } from '../../wailsjs/go/main/App';
    import { onMount } from 'svelte';
    import BlockUIElement from './BlockUIElement.svelte';

    let tasks: main.ScheduledTaskDto[] = [];
    let currentEditingId: string | null = null;
    let editingDocument: main.DocumentDto | null = null;
    let editingBlock: main.BlockDto | null = null;
    let loadingTaskId: string | null = null;

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
            .replace(/\/DONE/g, '')
            .trim();
        return withoutSchedule.replace(/(?:\r\n|\r|\n)/g, '<br>');
    }

    function requestEdit(id: string | null) {
        const wasEditing = currentEditingId !== null;
        currentEditingId = id;
        if (id === null) {
            editingDocument = null;
            editingBlock = null;
            if (wasEditing) {
                // Refresh tasks when exiting edit mode
                handleEditExit();
            }
        }
    }

    async function startEditing(task: main.ScheduledTaskDto) {
        if (loadingTaskId) return;
        
        loadingTaskId = task.block_id;
        try {
            const doc = await OpenDocument(task.doc_id);
            editingDocument = doc;
            editingBlock = doc.blocks.find((b: main.BlockDto) => b.id === task.block_id) || null;
            
            if (editingBlock) {
                currentEditingId = task.block_id;
            }
        } catch (err) {
            console.error('Failed to load document for editing:', err);
        } finally {
            loadingTaskId = null;
        }
    }

    async function handleSave() {
        if (!editingDocument || !editingBlock) return;
        
        try {
            // Update the block in the document
            const blockIndex = editingDocument.blocks.findIndex((b: main.BlockDto) => b.id === editingBlock!.id);
            if (blockIndex !== -1) {
                editingDocument.blocks[blockIndex] = editingBlock;
            }
            
            await SaveDocument(editingDocument);
            // Don't exit edit mode or refresh here - let the user continue editing
            // The task list will refresh when edit mode is exited
        } catch (err) {
            console.error('Failed to save:', err);
        }
    }

    async function handleEditExit() {
        // Refresh the tasks list when exiting edit mode to show updated description
        tasks = await GetScheduledTasks();
    }

    function handleDescriptionClick(task: main.ScheduledTaskDto) {
        startEditing(task);
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
                        {#if currentEditingId === task.block_id && editingBlock}
                            <div class="task-editor">
                                <BlockUIElement
                                    block={editingBlock}
                                    {currentEditingId}
                                    {requestEdit}
                                    on:save={handleSave}
                                />
                            </div>
                        {:else}
                            <div 
                                class="task-desc" 
                                class:loading={loadingTaskId === task.block_id}
                                role="button"
                                tabindex="0"
                                on:click={() => handleDescriptionClick(task)}
                                on:keydown={(e) => {
                                    if (e.key === 'Enter' || e.key === ' ') {
                                        e.preventDefault();
                                        handleDescriptionClick(task);
                                    }
                                }}
                            >
                                {#if loadingTaskId === task.block_id}
                                    <span class="loading-text">Loading...</span>
                                {:else}
                                    {@html cleanDescription(task.description)}
                                {/if}
                            </div>
                        {/if}
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
        transition: color 0.15s ease, background 0.15s ease;
        cursor: text;
        border-radius: 4px;
        padding: 2px 4px;
        margin-left: -4px;
    }

    .task-desc:hover {
        background: rgba(255, 255, 255, 0.06);
    }

    .task-desc:focus {
        outline: none;
        background: rgba(255, 255, 255, 0.06);
    }

    .task-desc.loading {
        cursor: wait;
        opacity: 0.7;
    }

    .loading-text {
        color: var(--text-dim);
        font-style: italic;
    }

    .task-editor {
        margin: 0;
        padding: 2px 0;
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