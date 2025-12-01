<script lang="ts">
  import { onMount } from "svelte";
  import type { Snippet } from "svelte";
  import "virtual:uno.css";
  import "../app.css";
  import ErrorBoundary from "$lib/components/ErrorBoundary.svelte";
  import ConnectionStatus from "$lib/components/ConnectionStatus.svelte";
  import { initPerformanceMonitoring } from "$lib/utils/performance";

  interface Props {
    children?: Snippet;
  }

  let { children }: Props = $props();

  onMount(() => {
    initPerformanceMonitoring();
  });
</script>

<ErrorBoundary context={{ component: "layout" }}>
  {@render children?.()}
  <ConnectionStatus />
</ErrorBoundary>
