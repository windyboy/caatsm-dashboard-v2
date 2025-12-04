/**
 * Utility functions for UI components
 */

/**
 * Merge class names (simple implementation)
 * Can be extended with clsx or similar library if needed
 */
export function cn(...classes: (string | undefined | null | false)[]): string {
  return classes.filter(Boolean).join(" ");
}

