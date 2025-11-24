<script lang="ts">
  import type { Telegram } from "../utils/types";

  export let telegram: Telegram;

  function getTypeColorClass(type: string): string {
    const typeLower = type.toLowerCase();
    if (typeLower === "aftn") {
      return "bg-sky-50/80 text-sky-700 group-hover:bg-sky-100/90";
    } else if (typeLower === "sita") {
      return "bg-rose-50/80 text-rose-700 group-hover:bg-rose-100/90";
    } else if (typeLower === "acars") {
      return "bg-emerald-50/80 text-emerald-700 group-hover:bg-emerald-100/90";
    } else if (typeLower === "cpdlc") {
      return "bg-violet-50/80 text-violet-700 group-hover:bg-violet-100/90";
    }
    return "bg-slate-50/80 text-slate-700 group-hover:bg-slate-100/90";
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

<div class="group rounded-lg bg-white/95 border-0 p-5 message-card-glow hover:bg-white/98 transition-all duration-200">
  <!-- Header -->
  <div class="flex justify-between items-start mb-4">
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-3 mb-2">
        <!-- Message ID -->
        <span class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-bold bg-slate-100/80 text-slate-800 border-0 shadow-sm group-hover:bg-slate-200/80 transition-colors">
          {telegram.message_id}
        </span>

        <!-- Flight Number -->
        {#if telegram.flight_number}
          <span class="text-xs text-slate-600 font-semibold group-hover:text-slate-800 transition-colors">
            {telegram.flight_number}
          </span>
        {/if}
      </div>
    </div>

    <!-- Time -->
    <span class="text-xs text-slate-500 whitespace-nowrap ml-3 font-mono font-medium group-hover:text-slate-700 transition-colors bg-slate-50/80 px-2.5 py-1 rounded-md border-0 shadow-sm">
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
    <span class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium shadow-sm border-0 transition-colors {getTypeColorClass(telegram.type)}">
      {telegram.type}
    </span>

    <!-- Priority -->
    <span class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium shadow-sm border-0 bg-purple-50/80 text-purple-700 group-hover:bg-purple-100/90 transition-colors">
      Priority {telegram.priority}
    </span>

    <!-- Route -->
    {#if telegram.source && telegram.destination}
      <span class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium shadow-sm border-0 bg-emerald-50/80 text-emerald-700 group-hover:bg-emerald-100/90 transition-colors">
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

