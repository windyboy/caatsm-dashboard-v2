import { vi } from "vitest";
import "@testing-library/jest-dom";

class MockWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;

  readyState = MockWebSocket.CONNECTING;
  url: string;
  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  private openTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(url: string) {
    this.url = url;
    // Use setTimeout - works with both real and fake timers
    this.openTimer = setTimeout(() => {
      this.readyState = MockWebSocket.OPEN;
      if (this.onopen) {
        this.onopen(new Event("open"));
      }
      this.openTimer = null;
    }, 10);
  }

  send(_data: string) {}

  close() {
    if (this.openTimer) {
      clearTimeout(this.openTimer);
      this.openTimer = null;
    }
    this.readyState = MockWebSocket.CLOSED;
    if (this.onclose) {
      this.onclose(new CloseEvent("close"));
    }
  }
}

(globalThis as any).WebSocket = MockWebSocket as any;

vi.mock("../src/lib/utils/logger.ts", () => ({
  createLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  }),
}));
