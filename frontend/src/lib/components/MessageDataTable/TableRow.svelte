<script lang="ts">
  import type { Telegram } from "$lib/types";
  import Badge from "../ui/Badge.svelte";
  import { formatPriorityShort, formatPriority, formatDate } from "$lib/utils/telegram-formatters";

  interface TableRowProps {
    message: Telegram;
  }

  let { message }: TableRowProps = $props();
</script>

<tr class="table-row" tabindex="0">
  <td class="table-cell">{formatDate(message.time, { format: "short" })}</td>
  <td class="table-cell">
    <Badge variant="soft">{message.type || "—"}</Badge>
  </td>
  <td class="table-cell table-cell--content">
    <span class="content-text" title={message.content || undefined}>
      {message.content || "No content"}
    </span>
  </td>
  <td class="table-cell">{message.source || "—"}</td>
  <td class="table-cell">{message.destination || "—"}</td>
  <td class="table-cell">{message.flight_number || "—"}</td>
  <td class="table-cell">
    <Badge
      variant={message.priority === 1 ? "error" : message.priority === 2 ? "warn" : "ghost"}
      title={formatPriority(message.priority)}
    >
      {formatPriorityShort(message.priority)}
    </Badge>
  </td>
</tr>

<style>
  .table-row {
    border-bottom: 1px solid #f4f4f5;
    transition: background-color 0.15s ease;
  }

  .table-row:hover {
    background: #fafafa;
  }

  .table-cell {
    padding: 0.75rem;
    color: #18181b;
    border-bottom: 1px solid #f4f4f5;
  }

  .table-cell--content {
    max-width: 400px;
  }

  .content-text {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  @media (max-width: 768px) {
    .table-cell {
      padding: 0.5rem;
    }
  }
</style>
