<script lang="ts">
  import type { Telegram } from "$lib/types";
  import Card from "./ui/Card.svelte";
  import Badge from "./ui/Badge.svelte";

  interface MessageListProps {
    messages?: Telegram[];
  }

  let { messages = [] }: MessageListProps = $props();

  const formatPriority = (priority?: number): string => {
    if (!priority) return "Priority N/A";
    if (priority === 1) return "Priority 1 · Urgent";
    if (priority === 2) return "Priority 2 · Operational";
    return "Priority 3 · Routine";
  };
</script>

<Card>
  <div class="card__header">
    <div>
      <p class="eyebrow">Latest Messages</p>
      <p class="muted">Newest first</p>
    </div>
    <Badge variant="soft">Live feed</Badge>
  </div>

  {#if messages.length === 0}
    <p class="muted" role="status" aria-live="polite">No messages yet.</p>
  {:else}
    <div class="message-list-container" role="region" aria-live="polite" aria-label="Latest messages">
      <ul class="message-list" role="list">
        {#each messages as message, index (`${message.message_id || "no-id"}-${index}`)}
          <li class="message" role="listitem">
            <div class="message-meta">
              <Badge variant="soft">{message.type || "Unknown"}</Badge>
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
              <Badge variant="ghost">{formatPriority(message.priority)}</Badge>
            </div>
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</Card>
