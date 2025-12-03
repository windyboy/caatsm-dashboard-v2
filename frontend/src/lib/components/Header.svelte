<script lang="ts">
  import ThemeToggle from "$lib/components/ThemeToggle.svelte";
  import { page } from "$app/stores";

  const currentPath = $derived($page.url.pathname);
  
  const isActive = (path: string) => currentPath === path;
</script>

<header
  class="bg-white/95 dark:bg-slate-800/95 backdrop-blur-lg shadow-lg border-b border-brand-200/30 dark:border-slate-700/50 sticky top-0 z-50"
>
  <div class="mx-auto max-w-7xl px-4 sm:px-6 py-4 sm:py-5">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2 sm:gap-4">
        <div
          class="w-8 h-8 sm:w-10 sm:h-10 rounded-lg bg-gradient-to-br from-brand-500 via-accent-500 to-success-500 shadow-lg border border-brand-400/30 flex items-center justify-center"
        >
          <svg
            class="w-4 h-4 sm:w-6 sm:h-6 text-white"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
            ></path>
          </svg>
        </div>
        <h1
          class="text-lg sm:text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight bg-gradient-to-r from-brand-600 to-accent-600 dark:from-brand-400 dark:to-accent-400 bg-clip-text text-transparent"
        >
          CAATSM Dashboard
        </h1>
      </div>
      <nav class="flex items-center gap-1 sm:gap-2">
        <a
          href="/"
          class="px-3 py-2 sm:px-4 sm:py-2.5 text-xs sm:text-sm font-semibold text-slate-700 dark:text-slate-300 hover:text-brand-600 dark:hover:text-brand-400 hover:bg-brand-50/80 dark:hover:bg-slate-700/80 rounded-lg transition-all duration-300 border {isActive('/') ? 'border-brand-300/60 dark:border-brand-500/60 bg-brand-50/50 dark:bg-brand-900/20 text-brand-600 dark:text-brand-400' : 'border-transparent hover:border-brand-200/60 dark:hover:border-slate-600/60'} hover:shadow-md hover:scale-105 flex items-center gap-1 sm:gap-2"
          aria-label="Dashboard"
        >
          <svg
            class="w-3 h-3 sm:w-4 sm:h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
            ></path>
          </svg>
          <span class="hidden xs:inline">Dashboard</span>
        </a>
        <a
          href="/search"
          class="px-3 py-2 sm:px-4 sm:py-2.5 text-xs sm:text-sm font-semibold text-slate-700 dark:text-slate-300 hover:text-brand-600 dark:hover:text-brand-400 hover:bg-brand-50/80 dark:hover:bg-slate-700/80 rounded-lg transition-all duration-300 border {isActive('/search') ? 'border-brand-300/60 dark:border-brand-500/60 bg-brand-50/50 dark:bg-brand-900/20 text-brand-600 dark:text-brand-400' : 'border-transparent hover:border-brand-200/60 dark:hover:border-slate-600/60'} hover:shadow-md hover:scale-105 flex items-center gap-1 sm:gap-2"
          aria-label="Search"
        >
          <svg
            class="w-3 h-3 sm:w-4 sm:h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            ></path>
          </svg>
          <span class="hidden xs:inline">Search</span>
        </a>
        <ThemeToggle />
      </nav>
    </div>
  </div>
</header>

