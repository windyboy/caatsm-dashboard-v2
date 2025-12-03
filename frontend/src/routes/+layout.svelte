<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import type { Snippet } from "svelte";
  import "virtual:uno.css";
  import "../app.css";
  import ErrorBoundary from "$lib/components/ErrorBoundary.svelte";
  import ConnectionStatus from "$lib/components/ConnectionStatus.svelte";
  import Header from "$lib/components/Header.svelte";
  import { initPerformanceMonitoring } from "$lib/utils/performance";
  import { websocket } from "$lib/stores/websocket";
  import { page } from "$app/stores";

  interface Props {
    children?: Snippet;
  }

  let { children }: Props = $props();

  const isLoginPage = $derived($page.url.pathname === "/login");

  onMount(() => {
    initPerformanceMonitoring();
    // Theme store already applies theme on initialization (see theme.ts)
    
    // Connect WebSocket at the layout level so it persists across page navigations
    // Only connect if not already connected
    if (typeof window !== "undefined" && !websocket.checkIsConnected()) {
      websocket.connect();
    }
  });

  onDestroy(() => {
    // Cleanup WebSocket when the entire app is destroyed (e.g., page unload)
    // This ensures proper cleanup on navigation away from the app
    if (typeof window !== "undefined") {
      websocket.cleanup();
    }
  });
</script>

<ErrorBoundary context={{ component: "layout" }}>
  {#if isLoginPage}
    {@render children?.()}
  {:else}
    <div
      class="min-h-screen text-slate-900 dark:text-slate-100 antialiased bg-gradient-to-br from-blue-50 via-blue-100 to-blue-200 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900"
      style="background-attachment: fixed;"
    >
      <Header />
      {@render children?.()}
    </div>
  {/if}
  <ConnectionStatus />
</ErrorBoundary>
