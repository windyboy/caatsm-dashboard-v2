<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import type { Snippet } from "svelte";
  import "virtual:uno.css";
  import "../app.css";
  import ErrorBoundary from "$lib/components/ErrorBoundary.svelte";
  import ConnectionStatus from "$lib/components/ConnectionStatus.svelte";
  import { initPerformanceMonitoring } from "$lib/utils/performance";
  import { theme } from "$lib/stores/theme";
  import { websocket } from "$lib/stores/websocket";

  interface Props {
    children?: Snippet;
  }

  let { children }: Props = $props();

  onMount(() => {
    initPerformanceMonitoring();
    // Initialize theme (store handles initialization, but ensure it's applied)
    theme.subscribe(() => {}); // Subscribe to ensure theme is applied
    
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
  {@render children?.()}
  <ConnectionStatus />
</ErrorBoundary>
