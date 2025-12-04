<script lang="ts">
  import { cn } from "./utils";
  import type { Snippet } from "svelte";

  interface ButtonProps {
    variant?: "default" | "ghost";
    disabled?: boolean;
    type?: "button" | "submit" | "reset";
    children?: Snippet;
    onclick?: (event: MouseEvent) => void;
    [key: string]: unknown;
  }

  let {
    variant = "default",
    disabled = false,
    type = "button",
    children,
    onclick,
    ...restProps
  }: ButtonProps = $props();
</script>

<button
  {type}
  {disabled}
  class={cn("button", variant === "ghost" && "button--ghost")}
  {onclick}
  {...restProps}
>
  {#if children}
    {@render children()}
  {/if}
</button>
