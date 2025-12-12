<script lang="ts">
  import { Badge as ShadcnBadge, type BadgeVariant } from "./badge";
  import type { Snippet } from "svelte";
  import { cn } from "$lib/utils";

  /**
   * Badge component wrapper that extends shadcn Badge with application-specific variants.
   * Provides semantic variants (ok, warn, error, idle) in addition to standard shadcn variants.
   * Use this component for consistent badge styling across the application.
   *
   * @example
   * ```svelte
   * <Badge variant="ok">Success</Badge>
   * <Badge variant="error">Error</Badge>
   * <Badge variant="soft">Info</Badge>
   * ```
   */
  interface BadgeProps {
    /** Badge variant - includes custom application variants (ok, warn, error, idle) and standard shadcn variants */
    variant?: "default" | "soft" | "ghost" | "ok" | "warn" | "error" | "idle";
    /** Additional CSS classes */
    class?: string;
    /** Tooltip title text */
    title?: string;
    /** Badge content */
    children?: Snippet;
  }

  let {
    variant = "default",
    class: className,
    title,
    children,
    ...restProps
  }: BadgeProps = $props();

  // Configuration object mapping custom variants to shadcn variants and custom classes
  const VARIANT_CONFIG: Record<
    NonNullable<BadgeProps["variant"]>,
    { shadcnVariant: BadgeVariant; customClasses?: string }
  > = {
    default: { shadcnVariant: "default" },
    soft: { shadcnVariant: "secondary" },
    ghost: { shadcnVariant: "outline" },
    ok: {
      shadcnVariant: "default",
      customClasses:
        "bg-green-50 text-green-700 border-green-200 dark:bg-green-950 dark:text-green-300 dark:border-green-800",
    },
    warn: {
      shadcnVariant: "secondary",
      customClasses:
        "bg-yellow-50 text-yellow-700 border-yellow-200 dark:bg-yellow-950 dark:text-yellow-300 dark:border-yellow-800",
    },
    error: { shadcnVariant: "destructive" },
    idle: {
      shadcnVariant: "outline",
      customClasses:
        "bg-gray-50 text-gray-600 border-gray-200 dark:bg-gray-950 dark:text-gray-400 dark:border-gray-800",
    },
  };

  const variantConfig = $derived(VARIANT_CONFIG[variant]);
  const shadcnVariant = $derived(variantConfig.shadcnVariant);
  const customClasses = $derived(variantConfig.customClasses ?? "");
</script>

<ShadcnBadge variant={shadcnVariant} class={cn(customClasses, className)} {title} {...restProps}>
  {#if children}
    {@render children()}
  {/if}
</ShadcnBadge>
