import { describe, it, expect, vi, afterEach } from "vitest";
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

const renderBoundary = (content = "Child content") =>
  render(ErrorBoundary as any, { slots: { default: content } } as any);

describe("ErrorBoundary", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("renders children when no error occurs", async () => {
    renderBoundary();

    await tick();
    const errorBoundary = document.querySelector(".error-boundary");
    expect(errorBoundary).not.toBeInTheDocument();
  });

  it("displays fallback message when error occurs", async () => {
    const { component } = renderBoundary();

    const testError = new Error("Test error");
    (component as any).handleError(testError);

    await tick();

    expect(screen.getByText("Oops! Something went wrong")).toBeInTheDocument();
    expect(screen.getByText("Something went wrong. Please try again.")).toBeInTheDocument();
  });

  it("retries and clears error when retry button is clicked", async () => {
    const { component } = renderBoundary();

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
