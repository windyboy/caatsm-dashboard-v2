<script lang="ts">
  import type { Telegram } from "$lib/types";
  import Badge from "./ui/Badge.svelte";
  import { formatPriority, formatDate } from "$lib/utils/telegram-formatters";

  interface MessageCardProps {
    message: Telegram;
  }

  let { message }: MessageCardProps = $props();
</script>

<li
  class="message p-6 bg-white border border-[#e4e4e7] rounded-xl transition-all shadow-sm hover:bg-[#fafafa] hover:border-[#d4d4d8] hover:shadow-md hover:-translate-y-px md:p-4"
>
  <div
    class="message-meta flex items-center gap-4 mb-4 flex-wrap md:flex-col md:items-start md:gap-2"
  >
    <Badge variant="soft">{message.type || "Unknown"}</Badge>
    <span class="muted">
      {formatDate(message.time)}
    </span>
  </div>
  <p class="message-body my-4 text-[#18181b] leading-relaxed break-words text-[0.95rem]">
    {message.content || "No content provided."}
  </p>
  <div
    class="message-foot flex items-center gap-6 flex-wrap mt-4 pt-4 border-t border-[#e4e4e7] md:flex-col md:items-start md:gap-2"
  >
    <span class="muted">
      {message.source || "----"} → {message.destination || "----"}
    </span>
    <span class="muted">{message.flight_number || "Flight N/A"}</span>
    <Badge variant="ghost">{formatPriority(message.priority)}</Badge>
  </div>
</li>
