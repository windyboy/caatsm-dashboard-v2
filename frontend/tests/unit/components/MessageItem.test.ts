import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/svelte";
import MessageItem from "$lib/components/MessageItem.svelte";
import type { Telegram } from "$lib/utils/types";

vi.mock("$lib/utils/logger", () => ({
  createLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  }),
}));

describe("MessageItem", () => {
  const mockTelegram: Telegram = {
    message_id: "MSG-12345",
    type: "AFTN",
    time: "2024-01-15T14:30:00Z",
    flight_number: "AA123",
    source: "KJFK",
    destination: "KLAX",
    priority: 1,
    content: "Test message content for display",
  };

  it("renders message data correctly", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    expect(screen.getByText("MSG-12345")).toBeInTheDocument();
    expect(screen.getByText("AA123")).toBeInTheDocument();
    expect(screen.getByText("Test message content for display")).toBeInTheDocument();
  });

  it("sanitizes XSS content", () => {
    const maliciousTelegram = {
      ...mockTelegram,
      content: '<script>alert("XSS")</script>Hello World',
    };

    render(MessageItem, {
      props: {
        telegram: maliciousTelegram,
      },
    });

    const contentElement = document.querySelector(".message-content");
    expect(contentElement?.textContent).toBe("Hello World");
    expect(document.querySelectorAll("script").length).toBe(0);
  });
});
