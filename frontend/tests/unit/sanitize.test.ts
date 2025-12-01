import { describe, it, expect } from "vitest";
import { sanitize } from "$lib/utils/sanitize";

describe("sanitize", () => {
  it("removes all HTML tags and preserves plain text", () => {
    expect(sanitize('<script>alert("XSS")</script>Hello')).toBe("Hello");
    expect(sanitize("<p>Paragraph</p><b>Bold</b>Text")).toBe("ParagraphBoldText");
    expect(sanitize('<img src="x" onerror="alert(\'XSS\')">Safe')).toBe("Safe");
    expect(sanitize('<iframe src="evil.com"></iframe>Content')).toBe("Content");
  });

  it("preserves plain text", () => {
    expect(sanitize("This is plain text content")).toBe("This is plain text content");
  });

  it("handles edge cases", () => {
    expect(sanitize("")).toBe("");
    expect(sanitize(null as any)).toBe("");
    expect(sanitize(undefined as any)).toBe("");
  });
});

