import DOMPurify from "isomorphic-dompurify";

/**
 * Sanitizes HTML content to prevent XSS attacks.
 * Currently, messages from the backend are trusted, but this utility
 * is available for future use cases where user input needs to be displayed.
 *
 * @param dirty - The potentially unsafe HTML string
 * @returns Sanitized HTML string safe for display
 */
export function sanitizeHtml(dirty: string): string {
  return DOMPurify.sanitize(dirty, {
    ALLOWED_TAGS: ["b", "i", "em", "strong", "a", "p", "br"],
    ALLOWED_ATTR: ["href", "target", "rel"],
  });
}

/**
 * Escapes HTML entities in plain text.
 *
 * @param text - The text to escape
 * @returns Escaped text safe for display
 */
export function escapeHtml(text: string): string {
  const map: Record<string, string> = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#039;",
  };
  return text.replace(/[&<>"']/g, (m) => map[m]);
}
