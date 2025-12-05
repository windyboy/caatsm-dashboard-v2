<script lang="ts">
  import { Card as ShadcnCard, CardContent } from "./card/index.js";
  import type { Snippet } from "svelte";

  interface CardProps {
    compact?: boolean;
    children?: Snippet;
    class?: string;
  }

  let { compact = false, children, class: className, ...restProps }: CardProps = $props();
</script>

<ShadcnCard class={className} {...restProps}>
  <CardContent class={compact ? "max-w-[520px] text-center" : ""}>
    {#if children}
      {@render children()}
    {/if}
  </CardContent>
</ShadcnCard>
