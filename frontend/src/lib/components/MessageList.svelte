<script lang="ts">
  import type { Telegram } from "$lib/types";

  export let messages: Telegram[] = [];

  const formatPriority = (priority?: number): string => {
    if (!priority) return "Priority N/A";
    if (priority === 1) return "Priority 1 · Urgent";
    if (priority === 2) return "Priority 2 · Operational";
    return "Priority 3 · Routine";
  };
</script>

<div class="card">
  <div class="card__header">
    <div>
      <p class="eyebrow">Latest Messages</p>
      <p class="muted">Newest first</p>
    </div>
    <span class="pill pill--soft">Live feed</span>
  </div>

  {#if messages.length === 0}
    <p class="muted">No messages yet.</p>
  {:else}
    <div class="message-list-container">
      <ul class="message-list">
        {#each messages as message, index (`${message.message_id || 'no-id'}-${index}`)}
          <li class="message">
            <div class="message-meta">
              <span class="pill pill--soft">{message.type || "Unknown"}</span>
              <span class="muted">
                {message.time ? new Date(message.time).toLocaleString() : "No timestamp"}
              </span>
            </div>

            <p class="message-body">{message.content || "No content provided."}</p>

            <div class="message-foot">
              <span class="muted">
                {message.source || "----"} → {message.destination || "----"}
              </span>
              <span class="muted">{message.flight_number || "Flight N/A"}</span>
              <span class="pill pill--ghost">{formatPriority(message.priority)}</span>
            </div>
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</div>
