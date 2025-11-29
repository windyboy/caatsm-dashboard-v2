<!--
  @component MessageItem
  Displays a single telegram message with formatted content and metadata.

  @param {Telegram} telegram - The telegram message to display
-->

<script lang="ts">
  import type { Telegram } from "../utils/types";

  interface Props {
    telegram: Telegram;
  }

  let { telegram }: Props = $props();

  function getTypeColorClass(type: string | undefined | null): string {
    // Handle undefined/null/empty type
    if (!type || typeof type !== "string") {
      return "bg-gradient-to-r from-slate-50 to-slate-100 text-slate-700 group-hover:from-slate-100 group-hover:to-slate-200 border-slate-200/50";
    }

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

  function formatTime(timeStr: string | undefined | null): string {
    if (!timeStr) {
      return "";
    }
    const date = new Date(timeStr);
    if (isNaN(date.getTime())) {
      return timeStr;
    }
    return date.toLocaleTimeString("en-US", { hour12: false });
  }
</script>

<div class="message-item group">
  <!-- Header -->
  <div class="flex justify-between items-start mb-4">
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-3 mb-2">
        <!-- Message ID -->
        <span class="message-id">
          {telegram.message_id || "N/A"}
        </span>

        <!-- Flight Number -->
        {#if telegram.flight_number}
          <span class="flight-number">
            {telegram.flight_number}
          </span>
        {/if}
      </div>
    </div>

    <!-- Time -->
    <span class="timestamp">
      {formatTime(telegram.time)}
    </span>
  </div>

  <!-- Message Content -->
  <p class="message-content">
    {telegram.content || ""}
  </p>

  <!-- Tags -->
  <div class="message-tags">
    <!-- Type Tag -->
    {#if telegram.type}
      <span class="tag type-tag {getTypeColorClass(telegram.type)}">
        {telegram.type}
      </span>
    {/if}

    <!-- Priority -->
    {#if telegram.priority}
      <span class="tag priority-tag">
        Priority {telegram.priority}
      </span>
    {/if}

    <!-- Route -->
    {#if telegram.source && telegram.destination}
      <span class="tag route-tag">
        {telegram.source} → {telegram.destination}
      </span>
    {/if}
  </div>
</div>

<style>
  .message-item {
    @apply rounded-lg p-5 transition-all duration-300 hover:shadow-lg hover:scale-[1.02];
    background: var(--message-card-bg);
    box-shadow: var(--message-card-shadow);
    border: var(--message-card-border);
  }

  .message-item:hover {
    background: var(--message-card-hover-bg);
    box-shadow: var(--message-card-hover-shadow);
  }

  .message-id {
    @apply inline-flex items-center px-2.5 py-1 rounded-md text-xs font-bold shadow-sm border;
    background: linear-gradient(to right, theme('colors.brand.100'), theme('colors.accent.100'));
    color: theme('colors.brand.800');
    border-color: theme('colors.brand.200 / 0.5');
    transition: all 0.3s;
  }

  .group:hover .message-id {
    background: linear-gradient(to right, theme('colors.brand.200'), theme('colors.accent.200'));
  }

  .flight-number {
    @apply text-xs font-semibold transition-colors;
    color: theme('colors.slate.600');
  }

  .group:hover .flight-number {
    color: theme('colors.slate.800');
  }

  .timestamp {
    @apply text-xs whitespace-nowrap ml-3 font-mono font-medium shadow-sm px-2.5 py-1 rounded-md border transition-colors;
    background: linear-gradient(to right, theme('colors.slate.50'), theme('colors.slate.100'));
    color: theme('colors.slate.600');
    border-color: theme('colors.slate.200 / 0.5');
  }

  .group:hover .timestamp {
    color: theme('colors.brand.700');
  }

  .message-content {
    @apply text-sm mb-3 break-words leading-relaxed;
    color: theme('colors.slate.800');
  }

  .message-tags {
    @apply flex flex-wrap gap-2;
  }

  .tag {
    @apply inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium shadow-sm border-0 transition-all duration-300;
  }

  .priority-tag {
    background: linear-gradient(to right, theme('colors.accent.50'), theme('colors.accent.100'));
    color: theme('colors.accent.700');
    border-color: theme('colors.accent.200 / 0.5');
  }

  .group:hover .priority-tag {
    background: linear-gradient(to right, theme('colors.accent.100'), theme('colors.accent.200'));
  }

  .route-tag {
    background: linear-gradient(to right, theme('colors.success.50'), theme('colors.success.100'));
    color: theme('colors.success.700');
    border-color: theme('colors.success.200 / 0.5');
  }

  .group:hover .route-tag {
    background: linear-gradient(to right, theme('colors.success.100'), theme('colors.success.200'));
  }

  /* Type tag colors - keeping the dynamic logic */
  .bg-gradient-to-r.from-brand-50.to-brand-100 {
    background: linear-gradient(to right, theme('colors.brand.50'), theme('colors.brand.100'));
    color: theme('colors.brand.700');
    border-color: theme('colors.brand.200 / 0.5');
  }

  .group:hover .bg-gradient-to-r.from-brand-50.to-brand-100 {
    background: linear-gradient(to right, theme('colors.brand.100'), theme('colors.brand.200'));
  }

  .bg-gradient-to-r.from-danger-50.to-warning-50 {
    background: linear-gradient(to right, theme('colors.danger.50'), theme('colors.warning.50'));
    color: theme('colors.danger.700');
    border-color: theme('colors.danger.200 / 0.5');
  }

  .group:hover .bg-gradient-to-r.from-danger-50.to-warning-50 {
    background: linear-gradient(to right, theme('colors.danger.100'), theme('colors.warning.100'));
  }

  .bg-gradient-to-r.from-success-50.to-brand-50 {
    background: linear-gradient(to right, theme('colors.success.50'), theme('colors.brand.50'));
    color: theme('colors.success.700');
    border-color: theme('colors.success.200 / 0.5');
  }

  .group:hover .bg-gradient-to-r.from-success-50.to-brand-50 {
    background: linear-gradient(to right, theme('colors.success.100'), theme('colors.brand.100'));
  }

  .bg-gradient-to-r.from-accent-50.to-brand-50 {
    background: linear-gradient(to right, theme('colors.accent.50'), theme('colors.brand.50'));
    color: theme('colors.accent.700');
    border-color: theme('colors.accent.200 / 0.5');
  }

  .group:hover .bg-gradient-to-r.from-accent-50.to-brand-50 {
    background: linear-gradient(to right, theme('colors.accent.100'), theme('colors.brand.100'));
  }

  .bg-gradient-to-r.from-slate-50.to-slate-100 {
    background: linear-gradient(to right, theme('colors.slate.50'), theme('colors.slate.100'));
    color: theme('colors.slate.700');
    border-color: theme('colors.slate.200 / 0.5');
  }

  .group:hover .bg-gradient-to-r.from-slate-50.to-slate-100 {
    background: linear-gradient(to right, theme('colors.slate.100'), theme('colors.slate.200'));
  }
</style>
