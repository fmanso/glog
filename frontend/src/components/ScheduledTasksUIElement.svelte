<script lang="ts">
    import { main } from '../../wailsjs/go/models';
    import { GetScheduledTasks } from '../../wailsjs/go/main/App';
    import { onMount } from 'svelte';
    let tasks: main.ScheduledTaskDto[] = [];

    onMount(async () => {
        tasks = await GetScheduledTasks();
    });
</script>

<main>
    {#if tasks }
        {#each tasks as task}
            <div class="task-item">
                <p>{task.description}</p>
            </div>
        {/each}
    {/if}
</main>

<style lang="css">

</style>