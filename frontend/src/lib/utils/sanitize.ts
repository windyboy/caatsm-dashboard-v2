import DOMPurify from "isomorphic-dompurify";

/**
 * Sanitizes HTML content to prevent XSS attacks.
 * Strips all HTML tags and returns plain text only.
 *
 * @param html - The HTML string to sanitize
 * @returns Sanitized plain text string
 */
export function sanitize(html: string): string {
  if (!html || typeof html !== "string") {
    return "";
  }

  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: [], // Allow no HTML tags (plain text only)
    ALLOWED_ATTR: [],
    KEEP_CONTENT: true, // Keep text content but strip tags
  });
}

