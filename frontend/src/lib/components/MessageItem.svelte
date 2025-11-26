<!--
  @component MessageItem
  Displays a single telegram message with formatted content and metadata.

  @param {Telegram} telegram - The telegram message to display
-->

<script lang="ts">
  import type { Telegram } from "../utils/types";

  /** @type {Telegram} */
  export let telegram: Telegram;

  function getTypeColorClass(type: string): string {
    const typeLower = type.toLowerCase();
    if (typeLower === "aftn") {
      return "bg-gradient-to-r from-brand-50 to-brand-100 text-brand-700 group-hover:from-brand-100 group-hover:to-brand-200 border-brand-200/50";
    } else if (typeLower === "sita") {
      return "bg-gradient-to-r from-danger-50 to-warning-50 text-danger-700 group-hover:from-danger-100 group-hover:to-warning-100 border-danger-200/50";
    } else if (typeLower === "acars") {
      return "bg-gradient-to-r from-success-50 to-brand-50 text-success-700 group-hover:from-success-100 group-hover:to-brand-100 border-success-200/50";
    } else if (typeLower === "cpdlc") {
      return "bg-gradient-to-r from-accent-50 to-brand-50 text-accent-700 group-hover:from-accent-100 group-hover:to-brand-100 border-accent-200/50";
    }
    return "bg-gradient-to-r from-slate-50 to-slate-100 text-slate-700 group-hover:from-slate-100 group-hover:to-slate-200 border-slate-200/50";
  }

  function formatTime(timeStr: string): string {
    try {
      const date = new Date(timeStr);
      return date.toLocaleTimeString("en-US", { hour12: false });
    } catch {
      return timeStr;
    }
  }
</script>

<div
  class="group rounded-lg bg-white/95 border-0 p-5 message-card-glow hover:bg-gradient-to-r hover:from-white hover:to-brand-50/50 transition-all duration-300 hover:shadow-lg hover:scale-[1.02]"
>
  <!-- Header -->
  <div class="flex justify-between items-start mb-4">
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-3 mb-2">
        <!-- Message ID -->
        <span
          class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-bold bg-gradient-to-r from-brand-100 to-accent-100 text-brand-800 border border-brand-200/50 shadow-sm group-hover:from-brand-200 group-hover:to-accent-200 transition-all duration-300"
        >
          {telegram.message_id}
        </span>

        <!-- Flight Number -->
        {#if telegram.flight_number}
          <span
            class="text-xs text-slate-600 font-semibold group-hover:text-slate-800 transition-colors"
          >
            {telegram.flight_number}
          </span>
        {/if}
      </div>
    </div>

    <!-- Time -->
    <span
      class="text-xs text-slate-600 whitespace-nowrap ml-3 font-mono font-medium group-hover:text-brand-700 transition-colors bg-gradient-to-r from-slate-50 to-slate-100 px-2.5 py-1 rounded-md border border-slate-200/50 shadow-sm"
    >
      {formatTime(telegram.time)}
    </span>
  </div>

  <!-- Message Content -->
  <p class="text-sm text-slate-800 mb-3 break-words leading-relaxed">
    {telegram.content}
  </p>

  <!-- Tags -->
  <div class="flex flex-wrap gap-2">
    <!-- Type Tag -->
    <span
      class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium shadow-sm border-0 transition-colors {getTypeColorClass(
        telegram.type
      )}"
    >
      {telegram.type}
    </span>

    <!-- Priority -->
    <span
      class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium shadow-sm border border-accent-200/50 bg-gradient-to-r from-accent-50 to-accent-100 text-accent-700 group-hover:from-accent-100 group-hover:to-accent-200 transition-all duration-300"
    >
      Priority {telegram.priority}
    </span>

    <!-- Route -->
    {#if telegram.source && telegram.destination}
      <span
        class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium shadow-sm border border-success-200/50 bg-gradient-to-r from-success-50 to-success-100 text-success-700 group-hover:from-success-100 group-hover:to-success-200 transition-all duration-300"
      >
        {telegram.source} → {telegram.destination}
      </span>
    {/if}
  </div>
</div>

<style>
  .message-card-glow {
    background: rgba(251, 252, 253, 0.95);
    backdrop-filter: blur(8px);
    box-shadow:
      0 2px 8px rgba(0, 0, 0, 0.03),
      0 1px 2px rgba(0, 0, 0, 0.02),
      0 0 0 0.5px rgba(0, 0, 0, 0.02),
      inset 0 1px 0 rgba(255, 255, 255, 0.95);
    border: 0.5px solid rgba(241, 245, 249, 0.6);
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .message-card-glow:hover {
    background: rgba(255, 255, 255, 0.98);
    box-shadow:
      0 4px 16px rgba(0, 0, 0, 0.06),
      0 2px 4px rgba(0, 0, 0, 0.03),
      0 0 0 0.5px rgba(0, 0, 0, 0.04),
      inset 0 1px 0 rgba(255, 255, 255, 1);
    border-color: rgba(226, 232, 240, 0.8);
    transform: translateY(-1px);
  }
</style>
