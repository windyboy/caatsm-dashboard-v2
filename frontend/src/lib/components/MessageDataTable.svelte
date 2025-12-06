<script lang="ts">
  import type { Telegram } from "$lib/types";
  import Card from "./ui/Card.svelte";
  import Badge from "./ui/Badge.svelte";
  import { Button } from "./ui/button/index.js";
  import { cn } from "$lib/utils.js";

  export type SortField = "time" | "type" | "priority" | "flight_number" | "source" | "destination";
  export type SortOrder = "asc" | "desc";

  interface MessageDataTableProps {
    messages?: Telegram[];
    pageSize?: number;
    loading?: boolean;
    onSort?: (field: SortField, order: SortOrder) => void;
    onPageChange?: (page: number) => void;
  }

  let {
    messages = [],
    pageSize = 10,
    loading = false,
    onSort,
    onPageChange,
  }: MessageDataTableProps = $props();

  let currentPage = $state(1);
  let sortField = $state<SortField>("time");
  let sortOrder = $state<SortOrder>("desc");

  const formatPriority = (priority?: number): string => {
    if (!priority) return "N/A";
    if (priority === 1) return "P1";
    if (priority === 2) return "P2";
    return "P3";
  };

  const getPriorityLabel = (priority?: number): string => {
    if (!priority) return "Priority N/A";
    if (priority === 1) return "Priority 1 · Urgent";
    if (priority === 2) return "Priority 2 · Operational";
    return "Priority 3 · Routine";
  };

  const formatDate = (dateStr?: string): string => {
    if (!dateStr) return "—";
    try {
      const date = new Date(dateStr);
      return date.toLocaleString("en-US", {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return dateStr;
    }
  };

  // Sort and paginate messages
  const sortedMessages = $derived.by(() => {
    // Only sort if we have messages to avoid unnecessary work
    if (messages.length === 0) return [];

    const sorted = [...messages];

    // Use stable sort for better performance and consistency
    sorted.sort((a, b) => {
      let aVal: string | number;
      let bVal: string | number;

      switch (sortField) {
        case "time":
          aVal = a.time ? new Date(a.time).getTime() : 0;
          bVal = b.time ? new Date(b.time).getTime() : 0;
          break;
        case "type":
          aVal = a.type || "";
          bVal = b.type || "";
          break;
        case "priority":
          aVal = a.priority || 0;
          bVal = b.priority || 0;
          break;
        case "flight_number":
          aVal = a.flight_number || "";
          bVal = b.flight_number || "";
          break;
        case "source":
          aVal = a.source || "";
          bVal = b.source || "";
          break;
        case "destination":
          aVal = a.destination || "";
          bVal = b.destination || "";
          break;
        default:
          return 0;
      }

      if (aVal < bVal) return sortOrder === "asc" ? -1 : 1;
      if (aVal > bVal) return sortOrder === "asc" ? 1 : -1;
      return 0;
    });

    return sorted;
  });

  const paginatedMessages = $derived.by(() => {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    return sortedMessages.slice(start, end);
  });

  const totalPages = $derived(Math.ceil(sortedMessages.length / pageSize));

  function handleSort(field: SortField) {
    if (sortField === field) {
      sortOrder = sortOrder === "asc" ? "desc" : "asc";
    } else {
      sortField = field;
      sortOrder = "desc";
    }
    onSort?.(sortField, sortOrder);
  }

  function handlePageChange(page: number) {
    currentPage = page;
    onPageChange?.(page);
  }

  const columns: Array<{
    key: SortField | "content";
    label: string;
    sortable: boolean;
    width?: string;
  }> = [
    { key: "time", label: "Time", sortable: true, width: "140px" },
    { key: "type", label: "Type", sortable: true, width: "100px" },
    { key: "content", label: "Content", sortable: false },
    { key: "source", label: "From", sortable: true, width: "100px" },
    { key: "destination", label: "To", sortable: true, width: "100px" },
    { key: "flight_number", label: "Flight", sortable: true, width: "100px" },
    { key: "priority", label: "Priority", sortable: true, width: "80px" },
  ];
</script>

<Card>
  <div class="data-table-container">
    <div class="card__header">
      <div>
        <p class="eyebrow">Latest Messages</p>
        <p class="muted">Real-time message stream</p>
      </div>
      <Badge variant="soft">Live feed</Badge>
    </div>

    {#if loading}
      <div class="table-loading" role="status" aria-live="polite" aria-busy="true">
        <p class="muted">Loading messages...</p>
      </div>
    {:else if paginatedMessages.length === 0}
      <div class="table-empty" role="status" aria-live="polite">
        <p class="muted">No messages available.</p>
      </div>
    {:else}
      <div class="table-wrapper" role="region" aria-label="Message data table">
        <table class="data-table">
          <thead>
            <tr>
              {#each columns as col (col.key)}
                <th
                  class={cn(
                    "table-header",
                    col.sortable && "table-header--sortable",
                    sortField === col.key && "table-header--active"
                  )}
                  style={col.width ? `width: ${col.width}` : undefined}
                  onclick={() => col.sortable && handleSort(col.key as SortField)}
                  onkeydown={(e) => {
                    if (col.sortable && (e.key === "Enter" || e.key === " ")) {
                      e.preventDefault();
                      handleSort(col.key as SortField);
                    }
                  }}
                  role="columnheader"
                  tabindex={col.sortable ? 0 : -1}
                  aria-sort={col.sortable && sortField === col.key
                    ? sortOrder === "asc"
                      ? "ascending"
                      : "descending"
                    : undefined}
                >
                  <div class="table-header-content">
                    <span>{col.label}</span>
                    {#if col.sortable && sortField === col.key}
                      <span class="sort-icon" aria-hidden="true">
                        {sortOrder === "asc" ? "↑" : "↓"}
                      </span>
                    {/if}
                  </div>
                </th>
              {/each}
            </tr>
          </thead>
          <tbody>
            {#each paginatedMessages as message, index (`${message.message_id || index}-${index}`)}
              <tr class="table-row" tabindex="0">
                <td class="table-cell">{formatDate(message.time)}</td>
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
                    variant={message.priority === 1
                      ? "error"
                      : message.priority === 2
                        ? "warn"
                        : "ghost"}
                    title={getPriorityLabel(message.priority)}
                  >
                    {formatPriority(message.priority)}
                  </Badge>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      {#if totalPages > 1}
        <div class="table-pagination">
          <Button
            variant="ghost"
            disabled={currentPage === 1}
            onclick={() => handlePageChange(currentPage - 1)}
            type="button"
          >
            Previous
          </Button>
          <span class="pagination-info">
            Page {currentPage} of {totalPages} ({sortedMessages.length} total)
          </span>
          <Button
            variant="ghost"
            disabled={currentPage >= totalPages}
            onclick={() => handlePageChange(currentPage + 1)}
            type="button"
          >
            Next
          </Button>
        </div>
      {/if}
    {/if}
  </div>
</Card>

<style>
  .data-table-container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .table-loading,
  .table-empty {
    padding: 2rem;
    text-align: center;
    color: #71717a;
  }

  .table-wrapper {
    overflow-x: auto;
    border: 1px solid #e4e4e7;
    border-radius: 8px;
    max-height: 600px;
    overflow-y: auto;
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .table-header {
    padding: 0.75rem;
    text-align: left;
    font-weight: 600;
    font-size: 0.75rem;
    color: #52525b;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid #e4e4e7;
    background: #fafafa;
    position: sticky;
    top: 0;
    z-index: 10;
  }

  .table-header--sortable {
    cursor: pointer;
    user-select: none;
    transition: background-color 0.15s ease;
  }

  .table-header--sortable:hover {
    background: #f4f4f5;
  }

  .table-header--active {
    color: #18181b;
  }

  .table-header-content {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .sort-icon {
    font-size: 0.875rem;
    color: #71717a;
  }

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

  .table-pagination {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e4e4e7;
    flex-wrap: wrap;
  }

  .pagination-info {
    font-size: 0.875rem;
    color: #71717a;
  }

  @media (max-width: 768px) {
    .table-wrapper {
      font-size: 0.75rem;
    }

    .table-header,
    .table-cell {
      padding: 0.5rem;
    }

    .table-pagination {
      flex-direction: column;
      align-items: stretch;
    }

    .pagination-info {
      text-align: center;
    }
  }
</style>
