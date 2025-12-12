<script lang="ts">
  import { cn } from "$lib/utils.js";
  import type { SortField, SortOrder } from "$lib/composables/useTableSort.svelte";

  interface TableHeaderProps {
    columns: Array<{
      key: SortField | "content";
      label: string;
      sortable: boolean;
      width?: string;
    }>;
    sortField: SortField;
    sortOrder: SortOrder;
    onSort: (field: SortField) => void;
  }

  let { columns, sortField, sortOrder, onSort }: TableHeaderProps = $props();
</script>

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
        onclick={() => col.sortable && onSort(col.key as SortField)}
        onkeydown={(e) => {
          if (col.sortable && (e.key === "Enter" || e.key === " ")) {
            e.preventDefault();
            onSort(col.key as SortField);
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

<style>
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

  @media (max-width: 768px) {
    .table-header {
      padding: 0.5rem;
    }
  }
</style>
