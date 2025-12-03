<script lang="ts">
  import "../app.css";
  import Header from "$lib/components/Header.svelte";
</script>

<div class="app-shell">
  <Header />
  <slot />
</div>
