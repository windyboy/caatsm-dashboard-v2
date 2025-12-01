import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import { tick } from "svelte";
import ErrorBoundary from "../../../src/lib/components/ErrorBoundary.svelte";

vi.mock("../../../src/lib/utils/logger", () => ({
  createLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  }),
}));

describe("ErrorBoundary", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("renders children when no error occurs", async () => {
    const TestChild = () => "Child content";

    render(ErrorBoundary, {
      props: {
        children: () => TestChild(),
      },
    });

    await tick();
    const errorBoundary = document.querySelector(".error-boundary");
    expect(errorBoundary).not.toBeInTheDocument();
  });

  it("displays fallback message when error occurs", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        children: () => "Child content",
      },
    });

    const testError = new Error("Test error");
    (component as any).handleError(testError);

    await tick();

    expect(screen.getByText("Oops! Something went wrong")).toBeInTheDocument();
    expect(screen.getByText("Something went wrong. Please try again.")).toBeInTheDocument();
  });

  it("retries and clears error when retry button is clicked", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        children: () => "Child content",
      },
    });

    (component as any).handleError(new Error("Test error"));
    await tick();

    expect(screen.getByText("Oops! Something went wrong")).toBeInTheDocument();

    const retryButton = screen.getByRole("button", { name: /try again/i });
    await fireEvent.click(retryButton);

    await tick();

    const errorBoundary = document.querySelector(".error-boundary");
    expect(errorBoundary).not.toBeInTheDocument();
  });
});
