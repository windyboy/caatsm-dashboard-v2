import { vi } from "vitest";
import "@testing-library/jest-dom/vitest";

// Mock logger to suppress console output during tests
vi.mock("$lib/utils/logger", () => ({
  createLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  }),
}));
