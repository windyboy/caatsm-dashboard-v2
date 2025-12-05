<script lang="ts">
  import { Badge as ShadcnBadge, type BadgeVariant } from "./badge/index.js";
  import type { Snippet } from "svelte";
  import { cn } from "$lib/utils.js";

  interface BadgeProps {
    variant?: "default" | "soft" | "ghost" | "ok" | "warn" | "error" | "idle";
    class?: string;
    title?: string;
    children?: Snippet;
  }

  let {
    variant = "default",
    class: className,
    title,
    children,
    ...restProps
  }: BadgeProps = $props();

  // Map custom variants to shadcn-svelte variants
  const mapVariant = (v: BadgeProps["variant"]): BadgeVariant => {
    switch (v) {
      case "soft":
        return "secondary";
      case "ghost":
        return "outline";
      case "ok":
        return "default";
      case "warn":
        return "secondary";
      case "error":
        return "destructive";
      case "idle":
        return "outline";
      default:
        return "default";
    }
  };

  // Custom classes for specific variants that need special styling
  const getCustomClasses = (v: BadgeProps["variant"]): string => {
    switch (v) {
      case "ok":
        return "bg-green-50 text-green-700 border-green-200 dark:bg-green-950 dark:text-green-300 dark:border-green-800";
      case "warn":
        return "bg-yellow-50 text-yellow-700 border-yellow-200 dark:bg-yellow-950 dark:text-yellow-300 dark:border-yellow-800";
      case "error":
        return ""; // Already handled by destructive variant
      case "idle":
        return "bg-gray-50 text-gray-600 border-gray-200 dark:bg-gray-950 dark:text-gray-400 dark:border-gray-800";
      default:
        return "";
    }
  };
</script>

<ShadcnBadge
  variant={mapVariant(variant)}
  class={cn(getCustomClasses(variant), className)}
  {title}
  {...restProps}
>
  {#if children}
    {@render children()}
  {/if}
</ShadcnBadge>
