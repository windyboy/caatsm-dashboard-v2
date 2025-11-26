<!--
  @component Button
  Reusable button component with multiple variants and states.

  @param {string} variant - Button style variant: 'primary', 'secondary', 'outline', 'ghost'
  @param {string} size - Button size: 'sm', 'md', 'lg'
  @param {boolean} disabled - Whether the button is disabled
  @param {boolean} loading - Whether to show loading state
  @param {boolean} ripple - Whether to enable ripple effect
  @param {string} type - Button type attribute
  @param {string} href - If provided, renders as anchor tag
-->

<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import LoadingSpinner from "./LoadingSpinner.svelte";

  const dispatch = createEventDispatcher();

  /** @type {'primary' | 'secondary' | 'outline' | 'ghost'} */
  export let variant: 'primary' | 'secondary' | 'outline' | 'ghost' = 'primary';
  /** @type {'sm' | 'md' | 'lg'} */
  export let size: 'sm' | 'md' | 'lg' = 'md';
  /** @type {boolean} */
  export let disabled: boolean = false;
  /** @type {boolean} */
  export let loading: boolean = false;
  /** @type {boolean} */
  export let ripple: boolean = true;
  /** @type {string} */
  export let type: string = 'button';
  /** @type {string | null} */
  export let href: string | null = null;

  const baseClasses = "inline-flex items-center justify-center font-semibold rounded-lg transition-all duration-300 focus:outline-none focus:ring-2 disabled:opacity-50 disabled:cursor-not-allowed";

  const variantClasses = {
    primary: "bg-gradient-to-r from-brand-500 to-accent-500 text-white hover:from-brand-600 hover:to-accent-600 focus:ring-brand-500/40 shadow-lg hover:shadow-xl",
    secondary: "bg-gradient-to-r from-slate-100 to-slate-200 text-slate-700 hover:from-slate-200 hover:to-slate-300 focus:ring-slate-500/40 shadow-sm hover:shadow-md",
    outline: "border-2 border-brand-500 text-brand-600 bg-transparent hover:bg-brand-50 focus:ring-brand-500/40",
    ghost: "text-slate-700 hover:text-brand-600 hover:bg-brand-50/80 focus:ring-brand-500/40"
  };

  const sizeClasses = {
    sm: "px-3 py-1.5 text-xs gap-1",
    md: "px-4 py-2 text-sm gap-2",
    lg: "px-6 py-3 text-base gap-3"
  };

  const interactionClasses = ripple ? "ripple btn-hover-lift" : "btn-hover-lift";

  $: classes = `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${interactionClasses}`;

  function handleClick(event: Event) {
    if (!disabled && !loading) {
      dispatch('click', event);
    }
  }
</script>

{#if href}
  <a {href} class={classes} on:click={handleClick}>
    {#if loading}
      <LoadingSpinner size="sm" variant="secondary" />
    {:else}
      <slot />
    {/if}
  </a>
{:else}
  <button {type} {disabled} class={classes} on:click={handleClick}>
    {#if loading}
      <LoadingSpinner size="sm" variant="secondary" />
    {:else}
      <slot />
    {/if}
  </button>
{/if}