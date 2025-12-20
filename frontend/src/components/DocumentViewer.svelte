<script lang="ts">
  import {main} from '../../wailsjs/go/models';
  export let document: main.DocumentDto;
</script>

<main>
    <h1>{document.title}</h1>

</main>


<style>
    h1 {
        font-size: 24px;
        margin: 0 0 16px 0;
    }

    main {
        background: #1e293b;
        padding: 24px;
        border-radius: 8px;
        margin-bottom: 24px;
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        justify-content: flex-start;
    }
</style>