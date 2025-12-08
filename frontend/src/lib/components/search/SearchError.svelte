<script lang="ts">
  interface SearchErrorProps {
    error: string;
    id?: string;
  }

  let { error, id = "search-error" }: SearchErrorProps = $props();
</script>

<div class="mb-6 animate-in fade-in duration-200">
  <div
    id={id}
    class="p-5 bg-red-50 border-l-4 border-red-500 rounded-lg shadow-sm"
    role="alert"
    aria-live="polite"
    aria-atomic="true"
  >
    <div class="flex items-start gap-3">
      <div class="flex-shrink-0">
        <svg class="w-5 h-5 text-red-600 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </div>
      <div class="flex-1">
        <h3 class="text-sm font-semibold text-red-900 mb-1">Error</h3>
        <p class="text-sm text-red-800">{error}</p>
      </div>
    </div>
  </div>
</div>

<style>
  @keyframes fade-in {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .animate-in {
    animation: fade-in 0.3s ease-out;
  }
</style>

