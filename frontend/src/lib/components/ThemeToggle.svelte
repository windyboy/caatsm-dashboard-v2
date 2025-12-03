<!--
  @component ThemeToggle
  Toggle button for switching between light and dark themes.
-->

<script lang="ts">
  import { theme } from "../stores/theme";
  import { onMount } from "svelte";

  let mounted = $state(false);
  let currentTheme = $state<"light" | "dark">("light");

  onMount(() => {
    mounted = true;
    const unsubscribe = theme.subscribe((t) => {
      currentTheme = t;
    });
    return unsubscribe;
  });
</script>

{#if mounted}
  <button
    type="button"
    onclick={() => theme.toggle()}
    class="theme-toggle"
    aria-label="Toggle theme"
    title={currentTheme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
  >
    {#if currentTheme === "dark"}
      <!-- Sun icon for light mode -->
      <svg
        class="w-5 h-5"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
        ></path>
      </svg>
    {:else}
      <!-- Moon icon for dark mode -->
      <svg
        class="w-5 h-5"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
        ></path>
      </svg>
    {/if}
  </button>
{/if}

<style>
  .theme-toggle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid transparent;
    background: transparent;
    color: currentColor;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .theme-toggle:hover {
    background: rgba(0, 0, 0, 0.05);
    border-color: rgba(0, 0, 0, 0.1);
  }

  :global(html.dark) .theme-toggle:hover {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.2);
  }

  .theme-toggle:focus {
    outline: none;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.4);
  }
</style>

