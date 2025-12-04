<script lang="ts">
  interface CardProps {
    compact?: boolean;
    children?: import("svelte").Snippet;
  }

  let { compact = false, children, ...restProps }: CardProps = $props();
</script>

<div class="card" class:card--compact={compact} {...restProps}>
  {#if children}
    {@render children()}
  {/if}
</div>
