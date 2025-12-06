<script lang="ts">
  import Input from "./ui/Input.svelte";
  import { Button } from "./ui/button/index.js";
  import Card from "./ui/Card.svelte";

  interface SearchFormProps {
    query?: string;
    startTime?: string;
    endTime?: string;
    loading?: boolean;
    onSubmit?: (event: SubmitEvent) => void;
    onReset?: () => void;
    onValidationError?: (error: string) => void;
  }

  let {
    query = $bindable(""),
    startTime = $bindable(""),
    endTime = $bindable(""),
    loading = false,
    onSubmit,
    onReset,
    onValidationError,
  }: SearchFormProps = $props();

  let showFilters = $state(false);

  function validateTimeRange(): string | null {
    if (startTime && endTime) {
      const start = new Date(startTime);
      const end = new Date(endTime);

      if (start >= end) {
        return "Start time must be before end time";
      }

      const daysDiff = (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24);
      if (daysDiff > 90) {
        return "Time range cannot exceed 90 days";
      }
    }
    return null;
  }

  function handleSubmit(event: SubmitEvent) {
    event.preventDefault();

    const validationError = validateTimeRange();
    if (validationError) {
      onValidationError?.(validationError);
      return;
    }

    onSubmit?.(event);
  }

  function handleReset() {
    query = "";
    startTime = "";
    endTime = "";
    showFilters = false;
    onReset?.();
  }
</script>

<Card class="shadow-lg border-slate-200">
  <form class="p-8 md:p-6" onsubmit={handleSubmit}>
    <!-- Main Search Bar -->
    <div class="flex gap-4 items-end mb-6 md:flex-col md:items-stretch">
      <div class="flex-1 min-w-0">
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
            <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <Input
            label="Search query"
            name="query"
            placeholder="Type, flight number, route, or text..."
            bind:value={query}
            ariaLabel="Search query"
            class="w-full pl-12"
          />
        </div>
      </div>
      <div class="flex gap-3 items-end md:w-full md:justify-between">
        <Button
          type="button"
          variant="ghost"
          onclick={() => showFilters = !showFilters}
          class="whitespace-nowrap border border-slate-300 hover:bg-slate-50"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
          </svg>
          {showFilters ? "Hide" : "Show"} Filters
        </Button>
        <div class="flex gap-2">
          <Button type="submit" disabled={loading} class="min-w-[120px] bg-blue-600 hover:bg-blue-700 text-white shadow-md hover:shadow-lg">
            {#if loading}
              <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
            {/if}
            {loading ? "Searching..." : "Search"}
          </Button>
          <Button
            type="button"
            variant="ghost"
            onclick={handleReset}
            disabled={loading}
            class="border border-slate-300 hover:bg-slate-50"
          >
            Clear
          </Button>
        </div>
      </div>
    </div>

    <!-- Expandable Filters Section -->
    {#if showFilters}
      <div
        class="pt-6 border-t border-slate-200 animate-in fade-in slide-in-from-top-2 duration-200"
      >
        <div class="mb-4">
          <h3 class="text-sm font-semibold text-slate-900 mb-4">Time Range Filters</h3>
        </div>
        <div class="grid grid-cols-2 gap-4 md:grid-cols-1">
          <Input
            label="Start Time"
            name="start_time"
            type="datetime-local"
            bind:value={startTime}
            ariaLabel="Start time"
          />
          <Input
            label="End Time"
            name="end_time"
            type="datetime-local"
            bind:value={endTime}
            ariaLabel="End time"
          />
        </div>
        {#if startTime || endTime}
          <div class="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <div class="flex items-start gap-2">
              <svg class="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p class="text-sm text-blue-900">
                <strong>Note:</strong> Time range is limited to 90 days maximum. Start time must be before end time.
              </p>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </form>
</Card>

<style>
  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes slide-in-from-top-2 {
    from {
      transform: translateY(-8px);
    }
    to {
      transform: translateY(0);
    }
  }

  .animate-in {
    animation: fade-in 0.2s ease-out, slide-in-from-top-2 0.2s ease-out;
  }
</style>

