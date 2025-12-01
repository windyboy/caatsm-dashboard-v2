<!--
  @component MessageItem
  Displays a single telegram message with formatted content and metadata.

  @param {Telegram} telegram - The telegram message to display
-->

<script lang="ts">
  import type { Telegram } from "../utils/types";
  import { sanitize } from "../utils/sanitize";

  interface Props {
    telegram: Telegram;
  }

  let { telegram }: Props = $props();

  const safeContent = $derived(sanitize(telegram.content || ""));

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
    {safeContent}
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
    border-radius: 0.5rem;
    padding: 1.25rem;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    transform: scale(1);
    background: var(--message-card-bg);
    box-shadow: var(--message-card-shadow);
    border: var(--message-card-border);
  }

  .message-item:hover {
    background: var(--message-card-hover-bg);
    box-shadow: var(--message-card-hover-shadow);
    transform: scale(1.02);
  }

  .message-id {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem 0.625rem;
    border-radius: 0.375rem;
    font-size: 0.75rem;
    line-height: 1rem;
    font-weight: 700;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    border: 1px solid transparent;
    background: linear-gradient(to right, #dbeafe, #f3e8ff);
    color: #1e40af;
    border-color: rgba(191, 219, 254, 0.5);
    transition: all 0.3s;
  }

  .group:hover .message-id {
    background: linear-gradient(to right, #bfdbfe, #e9d5ff);
  }

  .flight-number {
    font-size: 0.75rem;
    line-height: 1rem;
    font-weight: 600;
    transition: color 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    color: #475569;
  }

  .group:hover .flight-number {
    color: #1e293b;
  }

  .timestamp {
    font-size: 0.75rem;
    line-height: 1rem;
    white-space: nowrap;
    margin-left: 0.75rem;
    font-family:
      ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New",
      monospace;
    font-weight: 500;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    padding: 0.25rem 0.625rem;
    border-radius: 0.375rem;
    border: 1px solid transparent;
    transition: color 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    background: linear-gradient(to right, #f8fafc, #f1f5f9);
    color: #475569;
    border-color: rgba(226, 232, 240, 0.5);
  }

  .group:hover .timestamp {
    color: #1d4ed8;
  }

  .message-content {
    font-size: 0.875rem;
    line-height: 1.625;
    margin-bottom: 0.75rem;
    overflow-wrap: break-word;
    color: #1e293b;
  }

  .message-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .tag {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem 0.625rem;
    border-radius: 0.375rem;
    font-size: 0.75rem;
    line-height: 1rem;
    font-weight: 500;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    border: 1px solid transparent;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .priority-tag {
    background: linear-gradient(to right, #faf5ff, #f3e8ff);
    color: #7c3aed;
    border-color: rgba(233, 213, 255, 0.5);
  }

  .group:hover .priority-tag {
    background: linear-gradient(to right, #f3e8ff, #e9d5ff);
  }

  .route-tag {
    background: linear-gradient(to right, #f0fdf4, #dcfce7);
    color: #15803d;
    border-color: rgba(187, 247, 208, 0.5);
  }

  .group:hover .route-tag {
    background: linear-gradient(to right, #dcfce7, #bbf7d0);
  }

  /* Type tag colors - keeping the dynamic logic */
  .bg-gradient-to-r.from-brand-50.to-brand-100 {
    background: linear-gradient(to right, #eff6ff, #dbeafe);
    color: #1d4ed8;
    border-color: rgba(191, 219, 254, 0.5);
  }

  .group:hover .bg-gradient-to-r.from-brand-50.to-brand-100 {
    background: linear-gradient(to right, #dbeafe, #bfdbfe);
  }

  .bg-gradient-to-r.from-danger-50.to-warning-50 {
    background: linear-gradient(to right, #fef2f2, #fff7ed);
    color: #b91c1c;
    border-color: rgba(254, 202, 202, 0.5);
  }

  .group:hover .bg-gradient-to-r.from-danger-50.to-warning-50 {
    background: linear-gradient(to right, #fee2e2, #ffedd5);
  }

  .bg-gradient-to-r.from-success-50.to-brand-50 {
    background: linear-gradient(to right, #f0fdf4, #eff6ff);
    color: #15803d;
    border-color: rgba(187, 247, 208, 0.5);
  }

  .group:hover .bg-gradient-to-r.from-success-50.to-brand-50 {
    background: linear-gradient(to right, #dcfce7, #dbeafe);
  }

  .bg-gradient-to-r.from-accent-50.to-brand-50 {
    background: linear-gradient(to right, #faf5ff, #eff6ff);
    color: #7c3aed;
    border-color: rgba(233, 213, 255, 0.5);
  }

  .group:hover .bg-gradient-to-r.from-accent-50.to-brand-50 {
    background: linear-gradient(to right, #f3e8ff, #dbeafe);
  }

  .bg-gradient-to-r.from-slate-50.to-slate-100 {
    background: linear-gradient(to right, #f8fafc, #f1f5f9);
    color: #334155;
    border-color: rgba(226, 232, 240, 0.5);
  }

  .group:hover .bg-gradient-to-r.from-slate-50.to-slate-100 {
    background: linear-gradient(to right, #f1f5f9, #e2e8f0);
  }
</style>
