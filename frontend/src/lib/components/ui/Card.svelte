<script lang="ts">
  import { Card as ShadcnCard, CardContent } from "./card";
  import type { Snippet } from "svelte";
  import type { HTMLAttributes } from "svelte/elements";

  /**
   * Card component wrapper that provides a convenient API with automatic CardContent wrapping.
   * Use this component for simple card layouts. For more complex layouts with CardHeader,
   * CardFooter, etc., use the base Card components directly from "./card/index.js".
   *
   * @example
   * ```svelte
   * <Card>
   *   <p>Card content here</p>
   * </Card>
   * ```
   */
  interface CardProps extends Omit<HTMLAttributes<HTMLDivElement>, "class"> {
    /** Whether to use compact styling (centered, max-width constrained) */
    compact?: boolean;
    /** Card content */
    children?: Snippet;
    /** Additional CSS classes */
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
