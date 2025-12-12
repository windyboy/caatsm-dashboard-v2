<script lang="ts">
  import type { Telegram } from "$lib/types";
  import Card from "../ui/Card.svelte";
  import Badge from "../ui/Badge.svelte";
  import {
    useTableSort,
    type SortField,
    type SortOrder,
  } from "$lib/composables/useTableSort.svelte";
  import TableHeader from "./TableHeader.svelte";
  import TableRow from "./TableRow.svelte";
  import TablePagination from "./TablePagination.svelte";

  /**
   * Props for the MessageDataTable component.
   *
   * @interface MessageDataTableProps
   * @property {Telegram[]} [messages] - Array of telegram messages to display in the table
   * @property {number} [pageSize=10] - Number of messages to display per page
   * @property {boolean} [loading=false] - Whether the table is in a loading state
   * @property {(field: SortField, order: SortOrder) => void} [onSort] - Callback function called when sorting changes
   * @property {(page: number) => void} [onPageChange] - Callback function called when the page changes
   */
  interface MessageDataTableProps {
    /** Array of telegram messages to display in the table */
    messages?: Telegram[];
    /** Number of messages to display per page (default: 10) */
    pageSize?: number;
    /** Whether the table is in a loading state */
    loading?: boolean;
    /** Callback function called when sorting changes */
    onSort?: (field: SortField, order: SortOrder) => void;
    /** Callback function called when the page changes */
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

  const { sortField, sortOrder, sortedMessages, handleSort } = useTableSort({
    getMessages: () => messages,
    initialSortField: "time",
    initialSortOrder: "desc",
    getOnSort: () => onSort,
  });

  const paginatedMessages = $derived.by(() => {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    return sortedMessages.slice(start, end);
  });

  const totalPages = $derived(Math.ceil(sortedMessages.length / pageSize));

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

<svelte:boundary>
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
            <TableHeader {columns} {sortField} {sortOrder} onSort={handleSort} />
            <tbody>
              {#each paginatedMessages as message, index (`${message.message_id || index}-${index}`)}
                <TableRow {message} {index} />
              {/each}
            </tbody>
          </table>
        </div>

        <TablePagination
          {currentPage}
          {totalPages}
          totalItems={sortedMessages.length}
          onPageChange={handlePageChange}
        />
      {/if}
    </div>
  </Card>
</svelte:boundary>

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

  @media (max-width: 768px) {
    .table-wrapper {
      font-size: 0.75rem;
    }
  }
</style>
