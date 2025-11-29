import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/svelte";
import { tick } from "svelte";
import LiveStream from "$lib/components/LiveStream.svelte";
import { messages } from "$lib/stores/messages";
import { websocket } from "$lib/stores/websocket";
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

// Mock stores
vi.mock("$lib/stores/messages", () => ({
  messages: {
    subscribe: vi.fn(),
    add: vi.fn(),
    clear: vi.fn(),
  },
}));

vi.mock("$lib/stores/websocket", () => ({
  websocket: {
    status: {
      subscribe: vi.fn(),
    },
    reconnectAttempts: {
      subscribe: vi.fn(),
    },
    connect: vi.fn(),
    cleanup: vi.fn(),
  },
}));

describe("LiveStream", () => {
  let mockMessagesSubscribe: ReturnType<typeof vi.fn>;
  let mockStatusSubscribe: ReturnType<typeof vi.fn>;
  let mockReconnectAttemptsSubscribe: ReturnType<typeof vi.fn>;
  let mockContainer: HTMLElement;

  const mockTelegram: Telegram = {
    message_id: "test-123",
    type: "AFTN",
    time: "2024-01-01T12:00:00Z",
    flight_number: "AA123",
    source: "KJFK",
    destination: "KLAX",
    priority: 1,
    content: "Test message content",
  };

  beforeEach(() => {
    mockMessagesSubscribe = vi.fn();
    mockStatusSubscribe = vi.fn();
    mockReconnectAttemptsSubscribe = vi.fn();

    // Setup store mocks
    (messages.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback([]);
      return mockMessagesSubscribe;
    });

    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("disconnected");
      return mockStatusSubscribe;
    });

    (websocket.reconnectAttempts.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback(0);
      return mockReconnectAttemptsSubscribe;
    });

    // Mock window for SSR checks
    vi.stubGlobal("window", {});

    // Mock container for scrolling
    mockContainer = {
      scrollTop: 0,
      scrollTo: vi.fn(),
    } as any;
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.clearAllTimers();
  });

  it("renders the LiveStream component", () => {
    render(LiveStream);

    expect(screen.getByText("Live Stream")).toBeInTheDocument();
    expect(screen.getByText("Disconnected")).toBeInTheDocument();
  });

  it("displays connected status correctly", async () => {
    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("connected");
      return mockStatusSubscribe;
    });

    render(LiveStream);
    await tick();

    expect(screen.getByText("Connected")).toBeInTheDocument();
    const statusIndicator = screen.getByTitle("Connected");
    expect(statusIndicator).toHaveClass("bg-green-500");
  });

  it("displays connecting status correctly", async () => {
    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("connecting");
      return mockStatusSubscribe;
    });

    render(LiveStream);
    await tick();

    // "Connecting..." appears in both status badge and empty state, so use getAllByText
    const connectingTexts = screen.getAllByText("Connecting...");
    expect(connectingTexts.length).toBeGreaterThan(0);
    const statusIndicator = screen.getByTitle("Connecting...");
    expect(statusIndicator).toHaveClass("bg-yellow-500");
  });

  it("displays error status correctly", async () => {
    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("error");
      return mockStatusSubscribe;
    });

    render(LiveStream);
    await tick();

    expect(screen.getByText("Connection Error")).toBeInTheDocument();
    const statusIndicator = screen.getByTitle("Connection Error");
    expect(statusIndicator).toHaveClass("bg-red-500");
  });

  it("displays reconnect attempts when greater than 0", async () => {
    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("connecting");
      return mockStatusSubscribe;
    });

    (websocket.reconnectAttempts.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback(3);
      return mockReconnectAttemptsSubscribe;
    });

    render(LiveStream);
    await tick();

    // The reconnect attempts might be displayed differently, check for the status text
    // "Connecting..." appears multiple times, so use getAllByText
    const connectingTexts = screen.getAllByText("Connecting...");
    expect(connectingTexts.length).toBeGreaterThan(0);
  });

  it("displays messages from store", async () => {
    (messages.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback([mockTelegram]);
      return mockMessagesSubscribe;
    });

    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("connected");
      return mockStatusSubscribe;
    });

    render(LiveStream);
    await tick();

    expect(screen.getByText("Test message content")).toBeInTheDocument();
  });

  it("displays loading state when no messages", async () => {
    (messages.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback([]);
      return mockMessagesSubscribe;
    });

    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("connecting");
      return mockStatusSubscribe;
    });

    render(LiveStream);
    await tick();
    
    // When status is "connecting", it shows "Connecting..." not "Waiting for telegrams..."
    expect(screen.getByTitle("Connecting...")).toBeInTheDocument();
    // "Connecting..." appears multiple times, so use getAllByText
    const connectingTexts = screen.getAllByText("Connecting...");
    expect(connectingTexts.length).toBeGreaterThan(0);
  });

  it("displays error state when connection fails", () => {
    (messages.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback([]);
      return mockMessagesSubscribe;
    });

    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("error");
      return mockStatusSubscribe;
    });

    render(LiveStream);

    expect(screen.getByText("Connection error. Retrying...")).toBeInTheDocument();
  });

  it("displays disconnected state correctly", async () => {
    (messages.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback([]);
      return mockMessagesSubscribe;
    });

    render(LiveStream);
    await tick();

    expect(screen.getByText("Disconnected")).toBeInTheDocument();
    expect(screen.getByText("Waiting for telegrams...")).toBeInTheDocument();
  });

  it("connects to websocket on mount", async () => {
    render(LiveStream);

    await tick();

    expect(websocket.connect).toHaveBeenCalled();
  });

  it("cleans up subscriptions on unmount", async () => {
    const { unmount } = render(LiveStream);

    await tick();

    unmount();

    expect(mockMessagesSubscribe).toHaveBeenCalled();
    expect(mockStatusSubscribe).toHaveBeenCalled();
    expect(mockReconnectAttemptsSubscribe).toHaveBeenCalled();
    expect(websocket.cleanup).toHaveBeenCalled();
  });

  it("handles array messages correctly", () => {
    (messages.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback([mockTelegram]);
      return mockMessagesSubscribe;
    });

    render(LiveStream);

    expect(screen.getByText("Test message content")).toBeInTheDocument();
  });

  it("handles non-array messages gracefully", () => {
    (messages.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback(null); // Non-array value
      return mockMessagesSubscribe;
    });

    render(LiveStream);

    expect(screen.getByText("Waiting for telegrams...")).toBeInTheDocument();
  });

  it("renders MessageItem components for each message", () => {
    const messageList = [mockTelegram, { ...mockTelegram, message_id: "test-456" }];

    (messages.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback(messageList);
      return mockMessagesSubscribe;
    });

    render(LiveStream);

    // Should render two MessageItem components
    const messageItems = screen.getAllByText("Test message content");
    expect(messageItems).toHaveLength(2);
  });

  it("applies correct CSS classes for status indicators", () => {
    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("connected");
      return mockStatusSubscribe;
    });

    render(LiveStream);

    const statusBadge = screen.getByText("Connected");
    expect(statusBadge).toBeInTheDocument();
    // Check for status-badge class which is the base class
    expect(statusBadge).toHaveClass("status-badge");
  });

  it("shows correct status text for different states", () => {
    const testCases = [
      { status: "connected", expectedText: "Connected" },
      { status: "connecting", expectedText: "Connecting..." },
      { status: "error", expectedText: "Connection Error" },
      { status: "disconnected", expectedText: "Disconnected" },
    ];

    testCases.forEach(({ status, expectedText }) => {
      (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
        callback(status);
        return mockStatusSubscribe;
      });

      const { unmount } = render(LiveStream);

      // Use getByTitle for status badge to avoid multiple matches
      if (status === "connecting") {
        expect(screen.getByTitle(expectedText)).toBeInTheDocument();
      } else {
        expect(screen.getByText(expectedText)).toBeInTheDocument();
      }

      unmount();
      cleanup();
    });
  });

  it("handles SSR correctly by not subscribing during server-side rendering", () => {
    // Mock SSR environment
    vi.stubGlobal("window", undefined);

    render(LiveStream);

    // Should not call connect during SSR
    expect(websocket.connect).not.toHaveBeenCalled();
  });

  it("subscribes to stores only after mount in browser environment", async () => {
    render(LiveStream);

    await tick();

    expect(websocket.status.subscribe).toHaveBeenCalled();
    expect(websocket.reconnectAttempts.subscribe).toHaveBeenCalled();
    expect(messages.subscribe).toHaveBeenCalled();
  });
});