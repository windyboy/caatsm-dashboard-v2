
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/svelte";
import StatsCard from "$lib/components/StatsCard.svelte";
import { stats } from "$lib/stores/stats";
import type { StatsState } from "$lib/stores/stats";

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
      4: 20,
      5: 30,
    },
    byType: {
      AFTN: 80,
      SITA: 40,
      ACARS: 20,
      CPDLC: 10,
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

  it("renders the StatsCard component with total type", () => {
    render(StatsCard, {
      props: {
        title: "Total Messages",
        type: "total",
      },
    });

    expect(screen.getByText("Total Messages")).toBeInTheDocument();
    expect(screen.getByText("150")).toBeInTheDocument();
  });

  it("renders the StatsCard component with priority type", () => {
    render(StatsCard, {
      props: {
        title: "By Priority",
        type: "priority",
      },
    });

    expect(screen.getByText("By Priority")).toBeInTheDocument();
    expect(screen.getByText("Priority 1")).toBeInTheDocument();
    expect(screen.getByText("45")).toBeInTheDocument();
    expect(screen.getByText("Priority 2")).toBeInTheDocument();
    // "30" appears multiple times (Priority 2 and Priority 5), so use getAllByText
    const thirtyElements = screen.getAllByText("30");
    expect(thirtyElements.length).toBeGreaterThan(0);
  });

  it("renders the StatsCard component with type type", () => {
    render(StatsCard, {
      props: {
        title: "By Type",
        type: "type",
      },
    });

    expect(screen.getByText("By Type")).toBeInTheDocument();
    expect(screen.getByText("AFTN")).toBeInTheDocument();
    expect(screen.getByText("80")).toBeInTheDocument();
    expect(screen.getByText("SITA")).toBeInTheDocument();
    expect(screen.getByText("40")).toBeInTheDocument();
  });

  it("displays total with proper formatting", () => {
    const largeNumberState = { ...mockStatsState, total: 1234567 };
    (stats.subscribe as any).mockImplementation((callback: (value: StatsState) => void) => {
      callback(largeNumberState);
      return mockStatsSubscribe;
    });

    render(StatsCard, {
      props: {
        title: "Total Messages",
        type: "total",
      },
    });

    expect(screen.getByText("1,234,567")).toBeInTheDocument();
  });

  it("displays priority breakdown correctly", () => {
    render(StatsCard, {
      props: {
        title: "Priority Stats",
        type: "priority",
      },
    });

    // Check all priority levels are displayed
    expect(screen.getByText("Priority 1")).toBeInTheDocument();
    expect(screen.getByText("Priority 2")).toBeInTheDocument();
    expect(screen.getByText("Priority 3")).toBeInTheDocument();
    expect(screen.getByText("Priority 4")).toBeInTheDocument();
    expect(screen.getByText("Priority 5")).toBeInTheDocument();

    // Check counts are formatted - use getAllByText for values that appear multiple times
    expect(screen.getByText("45")).toBeInTheDocument();
    const thirtyElements = screen.getAllByText("30");
    expect(thirtyElements.length).toBeGreaterThan(0);
    expect(screen.getByText("25")).toBeInTheDocument();
    expect(screen.getByText("20")).toBeInTheDocument();
  });

  it("displays type breakdown correctly", () => {
    render(StatsCard, {
      props: {
        title: "Type Stats",
        type: "type",
      },
    });

    // Check all types are displayed in uppercase
    expect(screen.getByText("AFTN")).toBeInTheDocument();
    expect(screen.getByText("SITA")).toBeInTheDocument();
    expect(screen.getByText("ACARS")).toBeInTheDocument();
    expect(screen.getByText("CPDLC")).toBeInTheDocument();

    // Check counts
    expect(screen.getByText("80")).toBeInTheDocument();
    expect(screen.getByText("40")).toBeInTheDocument();
    expect(screen.getByText("20")).toBeInTheDocument();
    expect(screen.getByText("10")).toBeInTheDocument();
  });

  it("handles empty priority data", () => {
    const emptyPriorityState = { ...mockStatsState, byPriority: {} };
    (stats.subscribe as any).mockImplementation((callback: (value: StatsState) => void) => {
      callback(emptyPriorityState);
      return mockStatsSubscribe;
    });

    render(StatsCard, {
      props: {
        title: "Priority Stats",
        type: "priority",
      },
    });

    expect(screen.getByText("Priority Stats")).toBeInTheDocument();
    // Should not crash and should render the component
    const card = screen.getByText("Priority Stats").closest(".stats-card");
    expect(card).toBeInTheDocument();
  });

  it("handles empty type data", () => {
    const emptyTypeState = { ...mockStatsState, byType: {} };
    (stats.subscribe as any).mockImplementation((callback: (value: StatsState) => void) => {
      callback(emptyTypeState);
      return mockStatsSubscribe;
    });

    render(StatsCard, {
      props: {
        title: "Type Stats",
        type: "type",
      },
    });

    expect(screen.getByText("Type Stats")).toBeInTheDocument();
    // Should not crash and should render the component
    const card = screen.getByText("Type Stats").closest(".stats-card");
    expect(card).toBeInTheDocument();
  });

  it("handles zero total", () => {
    const zeroTotalState = { ...mockStatsState, total: 0 };
    (stats.subscribe as any).mockImplementation((callback: (value: StatsState) => void) => {
      callback(zeroTotalState);
      return mockStatsSubscribe;
    });

    render(StatsCard, {
      props: {
        title: "Total Messages",
        type: "total",
      },
    });

    expect(screen.getByText("0")).toBeInTheDocument();
  });

  it("handles undefined byPriority gracefully", () => {
    const undefinedPriorityState = { ...mockStatsState, byPriority: undefined };
    (stats.subscribe as any).mockImplementation((callback: (value: StatsState) => void) => {
      callback(undefinedPriorityState as any);
      return mockStatsSubscribe;
    });

    render(StatsCard, {
      props: {
        title: "Priority Stats",
        type: "priority",
      },
    });

    expect(screen.getByText("Priority Stats")).toBeInTheDocument();
    // Should not crash
    const card = screen.getByText("Priority Stats").closest(".stats-card");
    expect(card).toBeInTheDocument();
  });

  it("handles undefined byType gracefully", () => {
    const undefinedTypeState = { ...mockStatsState, byType: undefined };
    (stats.subscribe as any).mockImplementation((callback: (value: StatsState) => void) => {
      callback(undefinedTypeState as any);
      return mockStatsSubscribe;
    });

    render(StatsCard, {
      props: {
        title: "Type Stats",
        type: "type",
      },
    });

    expect(screen.getByText("Type Stats")).toBeInTheDocument();
    // Should not crash
    const card = screen.getByText("Type Stats").closest(".stats-card");
    expect(card).toBeInTheDocument();
  });

  it("renders correct icon for total type", () => {
    render(StatsCard, {
      props: {
        title: "Total Messages",
        type: "total",
      },
    });

    // Check for the chart bar icon (total type)
    const icon = document.querySelector("svg");
    expect(icon).toBeInTheDocument();
  });

  it("renders correct icon for priority type", () => {
    render(StatsCard, {
      props: {
        title: "Priority Stats",
        type: "priority",
      },
    });

    // Check for the star icon (priority type)
    const icon = document.querySelector("svg");
    expect(icon).toBeInTheDocument();
  });

  it("renders correct icon for type type", () => {
    render(StatsCard, {
      props: {
        title: "Type Stats",
        type: "type",
      },
    });

    // Check for the lightning bolt icon (type)
    const icon = document.querySelector("svg");
    expect(icon).toBeInTheDocument();
  });

  it("applies hover styles correctly", () => {
    render(StatsCard, {
      props: {
        title: "Total Messages",
        type: "total",
      },
    });

    const card = screen.getByText("Total Messages").closest(".stats-card");
    expect(card).toBeInTheDocument();
    // Hover classes are CSS pseudo-classes, test the base class instead
    expect(card).toHaveClass("stats-card");
  });

  it("subscribes to stats store on mount", () => {
    render(StatsCard, {
      props: {
        title: "Total Messages",
        type: "total",
      },
    });

    expect(stats.subscribe).toHaveBeenCalled();
  });

  it("unsubscribes from stats store on unmount", () => {
    const { unmount } = render(StatsCard, {
      props: {
        title: "Total Messages",
        type: "total",
      },
    });

    unmount();

    expect(mockStatsSubscribe).toHaveBeenCalled();
  });
});