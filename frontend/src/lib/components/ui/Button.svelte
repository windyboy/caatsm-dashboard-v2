<script lang="ts">
  import { Button as ShadcnButton, type ButtonProps as ShadcnButtonProps } from "./button/index.js";
  import type { Snippet } from "svelte";

  interface ButtonProps extends Omit<ShadcnButtonProps, "children"> {
    variant?: "default" | "ghost" | "outline" | "secondary" | "destructive" | "link";
    children?: Snippet;
    onclick?: (event: MouseEvent) => void;
  }

  let { variant = "default", children, onclick, ...restProps }: ButtonProps = $props();
</script>

<ShadcnButton {variant} {onclick} {...restProps}>
  {#if children}
    {@render children()}
  {/if}
</ShadcnButton>
