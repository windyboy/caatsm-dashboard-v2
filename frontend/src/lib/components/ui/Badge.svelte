<script lang="ts">
  import { cn } from "./utils";

  interface BadgeProps {
    variant?: "default" | "soft" | "ghost" | "ok" | "warn" | "error" | "idle";
    class?: string;
    title?: string;
    children?: import("svelte").Snippet;
  }

  let {
    variant = "default",
    class: className,
    title,
    children,
    ...restProps
  }: BadgeProps = $props();
</script>

<span
  class={cn(
    "pill",
    variant === "soft" && "pill--soft",
    variant === "ghost" && "pill--ghost",
    variant === "ok" && "pill--ok",
    variant === "warn" && "pill--warn",
    variant === "error" && "pill--error",
    variant === "idle" && "pill--idle",
    className
  )}
  {title}
  {...restProps}
>
  {#if children}
    {@render children()}
  {/if}
</span>
