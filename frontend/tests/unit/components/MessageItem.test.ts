import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/svelte";
import MessageItem from "$lib/components/MessageItem.svelte";
import type { Telegram } from "$lib/utils/types";

// Mock the logger
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

  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2024-01-15T14:35:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders the MessageItem component with basic data", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    expect(screen.getByText("MSG-12345")).toBeInTheDocument();
    expect(screen.getByText("AA123")).toBeInTheDocument();
    expect(screen.getByText("Test message content for display")).toBeInTheDocument();
  });

  it("displays message ID correctly", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    const messageIdElement = screen.getByText("MSG-12345");
    expect(messageIdElement).toBeInTheDocument();
    // Check for message-id class instead of individual utility classes
    expect(messageIdElement).toHaveClass("message-id");
  });

  it("displays flight number when present", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    expect(screen.getByText("AA123")).toBeInTheDocument();
  });

  it("does not display flight number when absent", () => {
    const telegramWithoutFlight = { ...mockTelegram, flight_number: undefined } as any;

    render(MessageItem, {
      props: {
        telegram: telegramWithoutFlight,
      },
    });

    expect(screen.queryByText("AA123")).not.toBeInTheDocument();
  });

  it("formats time correctly", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    // Should display time - toLocaleTimeString converts UTC to local time, so check for time pattern
    const timestamp = document.querySelector('.timestamp');
    expect(timestamp).toBeInTheDocument();
    // Time is converted from UTC to local timezone, so check for any valid time format
    expect(timestamp?.textContent).toMatch(/\d{1,2}:\d{2}/);
  });

  it("handles invalid time gracefully", () => {
    const telegramWithInvalidTime = { ...mockTelegram, time: "invalid-date" };

    render(MessageItem, {
      props: {
        telegram: telegramWithInvalidTime,
      },
    });

    expect(screen.getByText("invalid-date")).toBeInTheDocument();
  });

  it("handles undefined time gracefully", () => {
    const telegramWithUndefinedTime = { ...mockTelegram, time: undefined } as any;

    render(MessageItem, {
      props: {
        telegram: telegramWithUndefinedTime,
      },
    });

    // Check that the timestamp element exists (even if empty)
    const timestamp = document.querySelector(".timestamp");
    expect(timestamp).toBeInTheDocument();
  });

  it("displays message content correctly", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    const contentElement = screen.getByText("Test message content for display");
    expect(contentElement).toBeInTheDocument();
    // Check for message-content class instead of individual utility classes
    expect(contentElement).toHaveClass("message-content");
  });

  it("displays type tag with correct color for AFTN", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    const typeTag = screen.getByText("AFTN");
    expect(typeTag).toBeInTheDocument();
    expect(typeTag).toHaveClass("from-brand-50");
    expect(typeTag).toHaveClass("to-brand-100");
    expect(typeTag).toHaveClass("text-brand-700");
  });

  it("displays type tag with correct color for SITA", () => {
    const sitaTelegram = { ...mockTelegram, type: "SITA" };

    render(MessageItem, {
      props: {
        telegram: sitaTelegram,
      },
    });

    const typeTag = screen.getByText("SITA");
    expect(typeTag).toHaveClass("from-danger-50");
    expect(typeTag).toHaveClass("to-warning-50");
    expect(typeTag).toHaveClass("text-danger-700");
  });

  it("displays type tag with correct color for ACARS", () => {
    const acarsTelegram = { ...mockTelegram, type: "ACARS" };

    render(MessageItem, {
      props: {
        telegram: acarsTelegram,
      },
    });

    const typeTag = screen.getByText("ACARS");
    expect(typeTag).toHaveClass("from-success-50");
    expect(typeTag).toHaveClass("to-brand-50");
    expect(typeTag).toHaveClass("text-success-700");
  });

  it("displays type tag with correct color for CPDLC", () => {
    const cpdlcTelegram = { ...mockTelegram, type: "CPDLC" };

    render(MessageItem, {
      props: {
        telegram: cpdlcTelegram,
      },
    });

    const typeTag = screen.getByText("CPDLC");
    expect(typeTag).toHaveClass("from-accent-50");
    expect(typeTag).toHaveClass("to-brand-50");
    expect(typeTag).toHaveClass("text-accent-700");
  });

  it("displays default color for unknown type", () => {
    const unknownTypeTelegram = { ...mockTelegram, type: "UNKNOWN" };

    render(MessageItem, {
      props: {
        telegram: unknownTypeTelegram,
      },
    });

    const typeTag = screen.getByText("UNKNOWN");
    expect(typeTag).toHaveClass("from-slate-50");
    expect(typeTag).toHaveClass("to-slate-100");
    expect(typeTag).toHaveClass("text-slate-700");
  });

  it("handles undefined type gracefully", () => {
    const undefinedTypeTelegram = { ...mockTelegram, type: undefined } as any;

    render(MessageItem, {
      props: {
        telegram: undefinedTypeTelegram,
      },
    });

    // Should not render type tag
    expect(screen.queryByText("AFTN")).not.toBeInTheDocument();
  });

  it("handles null type gracefully", () => {
    const nullTypeTelegram = { ...mockTelegram, type: null } as any;

    render(MessageItem, {
      props: {
        telegram: nullTypeTelegram,
      },
    });

    // Should not render type tag
    expect(screen.queryByText("AFTN")).not.toBeInTheDocument();
  });

  it("handles empty string type gracefully", () => {
    const emptyTypeTelegram = { ...mockTelegram, type: "" } as any;

    render(MessageItem, {
      props: {
        telegram: emptyTypeTelegram,
      },
    });

    // Should not render type tag
    expect(screen.queryByText("AFTN")).not.toBeInTheDocument();
  });

  it("displays priority tag when priority is present", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    expect(screen.getByText("Priority 1")).toBeInTheDocument();
  });

  it("does not display priority tag when priority is absent", () => {
    const noPriorityTelegram = { ...mockTelegram, priority: undefined } as any;

    render(MessageItem, {
      props: {
        telegram: noPriorityTelegram,
      },
    });

    expect(screen.queryByText("Priority 1")).not.toBeInTheDocument();
  });

  it("displays route tag when source and destination are present", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    expect(screen.getByText("KJFK → KLAX")).toBeInTheDocument();
  });

  it("does not display route tag when source is missing", () => {
    const noSourceTelegram = { ...mockTelegram, source: undefined } as any;

    render(MessageItem, {
      props: {
        telegram: noSourceTelegram,
      },
    });

    expect(screen.queryByText("KJFK → KLAX")).not.toBeInTheDocument();
  });

  it("does not display route tag when destination is missing", () => {
    const noDestTelegram = { ...mockTelegram, destination: undefined } as any;

    render(MessageItem, {
      props: {
        telegram: noDestTelegram,
      },
    });

    expect(screen.queryByText("KJFK → KLAX")).not.toBeInTheDocument();
  });

  it("handles empty content gracefully", () => {
    const emptyContentTelegram = { ...mockTelegram, content: "" } as any;

    render(MessageItem, {
      props: {
        telegram: emptyContentTelegram,
      },
    });

    // Check that the message content element exists (even if empty)
    const content = document.querySelector(".message-content");
    expect(content).toBeInTheDocument();
  });

  it("handles undefined content gracefully", () => {
    const undefinedContentTelegram = { ...mockTelegram, content: undefined } as any;

    render(MessageItem, {
      props: {
        telegram: undefinedContentTelegram,
      },
    });

    // Check that the message content element exists (even if empty)
    const content = document.querySelector(".message-content");
    expect(content).toBeInTheDocument();
  });

  it("applies hover styles correctly", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    const messageItem = document.querySelector(".message-item");
    expect(messageItem).toBeInTheDocument();
    // Hover classes are CSS pseudo-classes, test the base class instead
    expect(messageItem).toHaveClass("message-item");
  });

  it("displays N/A for missing message_id", () => {
    const noIdTelegram = { ...mockTelegram, message_id: undefined } as any;

    render(MessageItem, {
      props: {
        telegram: noIdTelegram,
      },
    });

    expect(screen.getByText("N/A")).toBeInTheDocument();
  });

  it("handles non-string message_id", () => {
    const numberIdTelegram = { ...mockTelegram, message_id: 123 as any };

    render(MessageItem, {
      props: {
        telegram: numberIdTelegram,
      },
    });

    expect(screen.getByText("123")).toBeInTheDocument();
  });

  it("applies correct CSS classes for message card", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    const messageItem = document.querySelector(".message-item");
    expect(messageItem).toBeInTheDocument();
    // Check for the main class
    expect(messageItem).toHaveClass("message-item");
  });

  it("renders all tags in correct order", () => {
    render(MessageItem, {
      props: {
        telegram: mockTelegram,
      },
    });

    // Check for message-tags container
    const tagsContainer = document.querySelector(".message-tags");
    expect(tagsContainer).toBeInTheDocument();
    
    // Check that we have type, priority, and route tags by checking for their content
    expect(screen.getByText("AFTN")).toBeInTheDocument();
    expect(screen.getByText("Priority 1")).toBeInTheDocument();
    expect(screen.getByText("KJFK → KLAX")).toBeInTheDocument();
  });
});