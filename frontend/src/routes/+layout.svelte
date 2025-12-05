<script lang="ts">
  import "../app.css";
  import Header from "$lib/components/Header.svelte";
</script>

<a href="#main-content" class="skip-link">Skip to main content</a>
<div class="app-shell">
  <Header />
  <slot />
</div>

<style>
  .skip-link {
    position: absolute;
    top: -40px;
    left: 0;
    background: #000;
    color: #fff;
    padding: 8px;
    text-decoration: none;
    z-index: 100;
  }
  .skip-link:focus {
    top: 0;
  }
</style>
