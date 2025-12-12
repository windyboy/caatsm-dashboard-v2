import type { Telegram } from "$lib/types";

export type SortField = "time" | "type" | "priority" | "flight_number" | "source" | "destination";
export type SortOrder = "asc" | "desc";

export interface UseTableSortOptions {
  getMessages: () => Telegram[];
  initialSortField?: SortField;
  initialSortOrder?: SortOrder;
  getOnSort?: () => ((field: SortField, order: SortOrder) => void) | undefined;
}

export function useTableSort(options: UseTableSortOptions) {
  const { getMessages, initialSortField = "time", initialSortOrder = "desc", getOnSort } = options;

  let sortField = $state<SortField>(initialSortField);
  let sortOrder = $state<SortOrder>(initialSortOrder);

  const sortedMessages = $derived.by(() => {
    const messages = getMessages();
    if (messages.length === 0) return [];

    const sorted = [...messages];

    sorted.sort((a, b) => {
      let aVal: string | number;
      let bVal: string | number;

      switch (sortField) {
        case "time":
          // eslint-disable-next-line svelte/prefer-svelte-reactivity
          aVal = a.time ? new Date(a.time).getTime() : 0;
          // eslint-disable-next-line svelte/prefer-svelte-reactivity
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

  function handleSort(field: SortField) {
    if (sortField === field) {
      sortOrder = sortOrder === "asc" ? "desc" : "asc";
    } else {
      sortField = field;
      sortOrder = "desc";
    }
    getOnSort?.()?.(sortField, sortOrder);
  }

  return {
    sortField,
    sortOrder,
    sortedMessages,
    handleSort,
  };
}
