<!--
  Basic login page for authentication.
  This is a placeholder - actual authentication implementation depends on backend setup.
-->

<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { getWindow } from "$lib/utils/browser";

  let mounted = $state(false);
  let redirectPath = $state<string | null>(null);

  onMount(() => {
    mounted = true;
    // Get redirect path from sessionStorage
    const win = getWindow();
    if (win) {
      redirectPath = win.sessionStorage.getItem("redirectAfterLogin");
    }
  });

  function handleLogin() {
    // TODO: Implement actual authentication
    // For now, just redirect back
    const target = redirectPath || "/";
    const win = getWindow();
    if (win) {
      win.sessionStorage.removeItem("redirectAfterLogin");
    }
    goto(target);
  }
</script>

{#if mounted}
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
    <div class="max-w-md w-full mx-4">
      <div class="bg-white dark:bg-slate-800 rounded-lg shadow-lg p-8">
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white mb-6 text-center">
          CAATSM Dashboard
        </h1>
        <p class="text-sm text-slate-600 dark:text-slate-400 mb-6 text-center">
          Authentication required to access this resource.
        </p>
        <button
          type="button"
          onclick={handleLogin}
          class="w-full bg-brand-600 hover:bg-brand-700 text-white font-semibold py-2 px-4 rounded-lg transition-colors"
        >
          Login (Placeholder)
        </button>
        <p class="text-xs text-slate-500 dark:text-slate-500 mt-4 text-center">
          Note: This is a placeholder login page. Implement actual authentication based on your backend setup.
        </p>
      </div>
    </div>
  </div>
{/if}

