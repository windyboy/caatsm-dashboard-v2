import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/svelte";
import { tick } from "svelte";
import LiveStream from "$lib/components/LiveStream.svelte";
import { messages } from "$lib/stores/messages";
import { websocket } from "$lib/stores/websocket";
import type { Telegram } from "$lib/utils/types";

vi.mock("$lib/utils/logger", () => ({
  createLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  }),
}));

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
    checkIsConnected: vi.fn(() => false),
  },
}));

describe("LiveStream", () => {
  let mockMessagesSubscribe: ReturnType<typeof vi.fn>;
  let mockStatusSubscribe: ReturnType<typeof vi.fn>;
  let mockReconnectAttemptsSubscribe: ReturnType<typeof vi.fn>;

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

    (messages.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback([]);
      return mockMessagesSubscribe;
    });

    (websocket.status.subscribe as any).mockImplementation((callback: (value: any) => void) => {
      callback("disconnected");
      return mockStatusSubscribe;
    });

    (websocket.reconnectAttempts.subscribe as any).mockImplementation(
      (callback: (value: any) => void) => {
        callback(0);
        return mockReconnectAttemptsSubscribe;
      }
    );

    vi.stubGlobal("window", {});
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.clearAllTimers();
  });

  it("renders component with status", () => {
    render(LiveStream);

    expect(screen.getByText("Live Stream")).toBeInTheDocument();
    expect(screen.getByText("Disconnected")).toBeInTheDocument();
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
});
