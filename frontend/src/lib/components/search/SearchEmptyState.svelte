<script lang="ts">
  import Card from "../ui/Card.svelte";

  interface SearchEmptyStateProps {
    hasSearchCriteria: boolean;
    onReset?: () => void;
  }

  let { hasSearchCriteria, onReset }: SearchEmptyStateProps = $props();
</script>

{#if !hasSearchCriteria}
  <Card>
    <div class="py-10 px-6 text-center">
      <div class="max-w-md mx-auto">
        <div class="w-14 h-14 mx-auto mb-4 bg-blue-50 rounded-xl flex items-center justify-center border border-blue-100">
          <svg class="w-7 h-7 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
        <h2 class="text-base font-semibold text-slate-700 mb-2">
          Ready to search
        </h2>
        <p class="text-slate-500 mb-6 text-xs leading-relaxed">
          Enter a search query or use the time filters above to find messages in your database.
        </p>
        <div class="grid grid-cols-3 gap-2 max-w-xs mx-auto">
          <div class="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <div class="font-medium text-slate-700 mb-0.5 text-[11px]">Full-text</div>
            <div class="text-slate-500 text-[10px] leading-tight">Search content</div>
          </div>
          <div class="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <div class="font-medium text-slate-700 mb-0.5 text-[11px]">Time range</div>
            <div class="text-slate-500 text-[10px] leading-tight">Filter by date</div>
          </div>
          <div class="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <div class="font-medium text-slate-700 mb-0.5 text-[11px]">90 days</div>
            <div class="text-slate-500 text-[10px] leading-tight">Max range</div>
          </div>
        </div>
      </div>
    </div>
  </Card>
{:else}
  <Card>
    <div class="py-10 px-6 text-center">
      <div class="max-w-md mx-auto">
        <div class="w-14 h-14 mx-auto mb-4 bg-amber-50 rounded-xl flex items-center justify-center border border-amber-100">
          <svg class="w-7 h-7 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h2 class="text-base font-semibold text-slate-700 mb-2">
          No results found
        </h2>
        <p class="text-slate-500 mb-6 text-xs leading-relaxed">
          No messages match your search criteria. Try adjusting your filters or search terms.
        </p>
        {#if onReset}
          <button
            type="button"
            onclick={onReset}
            class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm hover:shadow-md transition-all"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            Clear search
          </button>
        {/if}
      </div>
    </div>
  </Card>
{/if}

