/**
 * Formatting utilities for Telegram messages
 */

/**
 * Formats priority as a full label (e.g., "Priority 1 · Urgent")
 */
export function formatPriority(priority?: number): string {
  if (!priority) return "Priority N/A";
  if (priority === 1) return "Priority 1 · Urgent";
  if (priority === 2) return "Priority 2 · Operational";
  return "Priority 3 · Routine";
}

/**
 * Formats priority as a short label (e.g., "P1")
 */
export function formatPriorityShort(priority?: number): string {
  if (!priority) return "N/A";
  if (priority === 1) return "P1";
  if (priority === 2) return "P2";
  return "P3";
}

/**
 * Formats a date string to a localized string
 * @param dateStr - ISO date string or undefined
 * @param options - Formatting options
 * @returns Formatted date string or fallback
 */
export function formatDate(
  dateStr?: string,
  options?: {
    format?: "full" | "short" | "time";
  }
): string {
  if (!dateStr) {
    if (options?.format === "time") return "—";
    return "No timestamp";
  }

  try {
    const date = new Date(dateStr);

    if (options?.format === "short") {
      return date.toLocaleString("en-US", {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    }

    if (options?.format === "time") {
      return date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
      });
    }

    return date.toLocaleString();
  } catch {
    return dateStr;
  }
}
