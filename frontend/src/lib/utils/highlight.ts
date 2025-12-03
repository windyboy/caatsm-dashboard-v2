import DOMPurify from "isomorphic-dompurify";

/**
 * Highlights search keywords in text by wrapping matches in <mark> tags.
 * Safe against XSS attacks by using DOMPurify for sanitization.
 *
 * @param text - The text to highlight keywords in
 * @param keywords - The search keywords to highlight (can be a single word or multiple words separated by spaces)
 * @returns HTML string with highlighted keywords wrapped in <mark> tags
 */
export function highlightText(text: string, keywords?: string | null): string {
  if (!text || typeof text !== "string") {
    return "";
  }

  if (!keywords || typeof keywords !== "string" || keywords.trim() === "") {
    return escapeHtml(text);
  }

  // Escape the original text to prevent XSS
  const escapedText = escapeHtml(text);

  // Split keywords by spaces and filter out empty strings
  const keywordList = keywords
    .trim()
    .split(/\s+/)
    .filter((kw) => kw.length > 0);

  if (keywordList.length === 0) {
    return escapedText;
  }

  // Create a regex pattern that matches any of the keywords (case-insensitive)
  // Escape special regex characters in keywords
  const escapedKeywords = keywordList.map((kw) => escapeRegex(kw));
  const pattern = new RegExp(`(${escapedKeywords.join("|")})`, "gi");

  // Replace matches with highlighted version
  const highlighted = escapedText.replace(pattern, '<mark class="highlight">$1</mark>');

  // Sanitize the result to ensure only allowed tags (mark) are present
  return DOMPurify.sanitize(highlighted, {
    ALLOWED_TAGS: ["mark"],
    ALLOWED_ATTR: ["class"],
  });
}

/**
 * Escapes HTML special characters to prevent XSS attacks.
 *
 * @param text - The text to escape
 * @returns Escaped HTML string
 */
function escapeHtml(text: string): string {
  const map: Record<string, string> = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#039;",
  };
  return text.replace(/[&<>"']/g, (m) => map[m]);
}

/**
 * Escapes special regex characters in a string.
 *
 * @param text - The text to escape for use in regex
 * @returns Escaped regex string
 */
function escapeRegex(text: string): string {
  return text.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

