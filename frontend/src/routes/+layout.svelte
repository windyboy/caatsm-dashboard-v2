<script lang="ts">
  import type { Snippet } from "svelte";
  import "virtual:uno.css";
  import "../app.css";
  import ErrorBoundary from "$lib/components/ErrorBoundary.svelte";

  interface Props {
    children?: Snippet;
  }

  let { children }: Props = $props();
</script>

<ErrorBoundary context={{ component: 'layout' }}>
  {@render children?.()}
</ErrorBoundary>
