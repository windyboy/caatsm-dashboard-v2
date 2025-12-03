<!--
  @component SearchResults
  Displays search results for telegram messages.

  @param {Telegram[] | null} [telegrams] - Array of telegram messages to display (optional, defaults to empty array if null/undefined)
  @param {number} [total] - Total number of results available (optional, defaults to 0 if undefined)
-->

<script lang="ts">
  import MessageItem from "./MessageItem.svelte";
  import type { Telegram } from "../utils/types";
  import VirtualList from "@sveltejs/svelte-virtual-list";

  interface Props {
    telegrams?: Telegram[] | null;
    total?: number;
  }

  let { telegrams = [], total = 0 }: Props = $props();

  // Ensure telegrams is always an array, never null
  const safeTelegrams = $derived(telegrams ?? []);

  // Generate a stable unique key for each telegram
  // Uses message_id if available, otherwise creates a composite key from stable fields
  // Keys must identify items, not positions, so we never use array indices
  function getTelegramKey(telegram: Telegram): string {
    // Primary: use message_id if it exists and is non-empty
    if (telegram.message_id && telegram.message_id.trim() !== "") {
      return telegram.message_id;
    }

    // Fallback: create stable composite key from immutable fields
    // time is always present (validated in domain), so this should be unique
    // We use a combination of fields that together uniquely identify the telegram
    const parts = [
      telegram.time || "",
      telegram.flight_number || "",
      telegram.source || "",
      telegram.destination || "",
      telegram.type || "",
      telegram.priority?.toString() || "",
      // Include content hash as additional differentiator for edge cases
      // This ensures uniqueness even if all other fields are identical
      telegram.content ? String(telegram.content.length) + telegram.content.substring(0, 20) : "",
    ];

    return parts.join("|");
  }
</script>

<div class="rounded-lg bg-white/95 backdrop-blur-md border-0 p-8 card-glow min-h-[400px]">
  {#if safeTelegrams.length > 0}
    <div class="space-y-5">
      <p class="text-base font-semibold text-slate-700 mb-5">
        Found <span
          class="bg-linear-to-r from-brand-600 to-accent-600 bg-clip-text text-transparent font-bold"
          >{total.toLocaleString()}</span
        > results
      </p>
      <VirtualList
        items={safeTelegrams}
        itemHeight={200}
        height="600px"
        class="virtual-list-container"
        let:item={telegram}
      >
        <div class="result-item">
          <MessageItem {telegram} />
        </div>
      </VirtualList>
    </div>
  {:else}
    <div class="flex flex-col items-center justify-center gap-2 text-slate-400 py-12">
      <div
        class="w-16 h-16 mx-auto mb-4 rounded-full bg-linear-to-r from-slate-200 to-slate-300"
      ></div>
      <p class="text-sm font-medium text-slate-500">No results found</p>
      <p class="text-xs text-slate-400">Try adjusting your search criteria</p>
    </div>
  {/if}
</div>

<style>
  .card-glow {
    background: rgba(252, 252, 253, 0.9);
    backdrop-filter: blur(10px);
    box-shadow:
      0 4px 16px rgba(0, 0, 0, 0.04),
      0 2px 4px rgba(0, 0, 0, 0.02),
      0 0 0 0.5px rgba(0, 0, 0, 0.03),
      inset 0 1px 0 rgba(255, 255, 255, 0.9);
    border: 0.5px solid rgba(226, 232, 240, 0.5);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .result-item {
    margin-bottom: 1rem;
  }

  .result-item:last-child {
    margin-bottom: 0;
  }
</style>
