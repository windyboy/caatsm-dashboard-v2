import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/svelte";
import StatsCard from "$lib/components/StatsCard.svelte";
import { stats } from "$lib/stores/stats";
import type { StatsState } from "$lib/stores/stats";

vi.mock("$lib/utils/logger", () => ({
  createLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  }),
}));

vi.mock("$lib/stores/stats", () => ({
  stats: {
    subscribe: vi.fn(),
  },
}));

describe("StatsCard", () => {
  let mockStatsSubscribe: ReturnType<typeof vi.fn>;

  const mockStatsState: StatsState = {
    total: 150,
    byPriority: {
      1: 45,
      2: 30,
      3: 25,
    },
    byType: {
      AFTN: 80,
      SITA: 40,
      ACARS: 30,
    },
  };

  beforeEach(() => {
    mockStatsSubscribe = vi.fn();
    (stats.subscribe as any).mockImplementation((callback: (value: StatsState) => void) => {
      callback(mockStatsState);
      return mockStatsSubscribe;
    });
    (stats as any).snapshot = () => mockStatsState;
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("renders stats", () => {
    render(StatsCard, {
      props: {
        title: "Total Messages",
        type: "total",
      },
    });

    expect(screen.getByText("Total Messages")).toBeInTheDocument();
    expect(screen.getByText("150")).toBeInTheDocument();
  });
});
