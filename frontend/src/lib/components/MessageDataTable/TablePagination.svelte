<script lang="ts">
  import { Button } from "../ui/button";

  interface TablePaginationProps {
    currentPage: number;
    totalPages: number;
    totalItems: number;
    onPageChange: (page: number) => void;
  }

  let { currentPage, totalPages, totalItems, onPageChange }: TablePaginationProps = $props();
</script>

{#if totalPages > 1}
  <div class="table-pagination">
    <Button
      variant="ghost"
      disabled={currentPage === 1}
      onclick={() => onPageChange(currentPage - 1)}
      type="button"
    >
      Previous
    </Button>
    <span class="pagination-info">
      Page {currentPage} of {totalPages} ({totalItems} total)
    </span>
    <Button
      variant="ghost"
      disabled={currentPage >= totalPages}
      onclick={() => onPageChange(currentPage + 1)}
      type="button"
    >
      Next
    </Button>
  </div>
{/if}

<style>
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
    .table-pagination {
      flex-direction: column;
      align-items: stretch;
    }

    .pagination-info {
      text-align: center;
    }
  }
</style>
