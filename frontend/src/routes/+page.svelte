<script lang="ts">
  import LiveStream from "$lib/components/LiveStream.svelte";
  import StatsCard from "$lib/components/StatsCard.svelte";
</script>

<div class="min-h-screen text-slate-900 antialiased" style="background: linear-gradient(135deg, rgb(238, 242, 247) 0%, rgb(235, 239, 245) 50%, rgb(237, 241, 246) 100%); background-attachment: fixed;">

  <header class="bg-white/95 backdrop-blur-lg shadow-sm border-b border-slate-200/60 sticky top-0 z-50">
    <div class="mx-auto max-w-7xl px-6 py-5">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-sky-500 to-blue-600 shadow-md border border-sky-600/20"></div>
          <h1 class="text-2xl font-bold text-slate-900 tracking-tight">
            CAATSM Dashboard
          </h1>
        </div>
        <nav class="flex items-center gap-2">
          <a href="/" class="px-4 py-2.5 text-sm font-semibold text-slate-700 hover:text-sky-600 hover:bg-sky-50/80 rounded-lg transition-all duration-200 border border-transparent hover:border-sky-200/60">
            Dashboard
          </a>
          <a href="/search" class="px-4 py-2.5 text-sm font-semibold text-slate-700 hover:text-sky-600 hover:bg-sky-50/80 rounded-lg transition-all duration-200 border border-transparent hover:border-sky-200/60">
            Search
          </a>
        </nav>
      </div>
    </div>
  </header>

  <main class="mx-auto max-w-7xl px-6 py-8">
    <section class="space-y-8">
      <div class="grid gap-6 lg:grid-cols-3">
        <div class="lg:col-span-2">
          <LiveStream />
        </div>
        <div class="space-y-4">
          <StatsCard title="Total Messages (24h)" type="total" />
          <StatsCard title="Priority Breakdown" type="priority" />
          <StatsCard title="Type Breakdown" type="type" />
        </div>
      </div>
    </section>
  </main>
</div>

