<script lang="ts">
  // CRITICAL: First line import, ensures init() executes before Header renders
  import '$lib/i18n';
  
  import { locale } from 'svelte-i18n';
  import "../app.css";
  import Header from "$lib/components/Header.svelte";

  export let data;

  // Update locale if server provided a different one
  // This happens AFTER initial render, so hydration is safe
  $: if (data?.locale) {
    $locale = data.locale;
  }
</script>

<a href="#main-content" class="skip-link">Skip to main content</a>
<div class="app-shell">
  <Header />
  <main id="main-content">
    <slot />
  </main>
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
