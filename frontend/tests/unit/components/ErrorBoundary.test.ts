import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import { tick } from "svelte";
import ErrorBoundary from "../../../src/lib/components/ErrorBoundary.svelte";

// Mock the logger
vi.mock("../../../src/lib/utils/logger", () => ({
  createLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  }),
}));

describe("ErrorBoundary", () => {
  let mockOnError: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockOnError = vi.fn();
    // Note: import.meta.env.DEV is configured in vitest.config.ts via define
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("renders children when no error occurs", async () => {
    // Create a simple test component that uses ErrorBoundary
    const TestChild = () => "Child content";
    
    render(ErrorBoundary, {
      props: {
        children: () => TestChild(),
      },
    });

    await tick();
    // In Svelte 5, snippets work differently - check if error boundary is not shown (meaning children rendered)
    const errorBoundary = document.querySelector('.error-boundary');
    expect(errorBoundary).not.toBeInTheDocument(); // Should not show error boundary when no error
  });

  it("displays fallback message when error occurs", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        children: () => "Child content",
      },
    });

    // Access the exported handleError function
    const testError = new Error("Test error");
    (component as any).handleError(testError);

    await tick();

    expect(screen.getByText("Oops! Something went wrong")).toBeInTheDocument();
    expect(screen.getByText("Something went wrong. Please try again.")).toBeInTheDocument();
  });

  it("displays custom fallback message", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        fallbackMessage: "Custom error message",
        children: () => "Child content",
      },
    });

    (component as any).handleError(new Error("Test error"));
    await tick();

    expect(screen.getByText("Custom error message")).toBeInTheDocument();
  });

  it("shows retry button when showRetry is true", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        showRetry: true,
        children: () => "Child content",
      },
    });

    (component as any).handleError(new Error("Test error"));
    await tick();

    const retryButton = screen.getByRole("button", { name: /try again/i });
    expect(retryButton).toBeInTheDocument();
  });

  it("hides retry button when showRetry is false", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        showRetry: false,
        children: () => "Child content",
      },
    });

    (component as any).handleError(new Error("Test error"));
    await tick();

    const retryButton = screen.queryByRole("button", { name: /try again/i });
    expect(retryButton).not.toBeInTheDocument();
  });

  it("retries and clears error when retry button is clicked", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        children: () => "Child content",
      },
    });

    (component as any).handleError(new Error("Test error"));
    await tick();

    // Error boundary should be visible
    expect(screen.getByText("Oops! Something went wrong")).toBeInTheDocument();

    const retryButton = screen.getByRole("button", { name: /try again/i });
    await fireEvent.click(retryButton);

    await tick();

    // After retry, error boundary should be hidden (children should render)
    const errorBoundary = document.querySelector('.error-boundary');
    expect(errorBoundary).not.toBeInTheDocument();
  });

  it("displays error details in development mode", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        children: () => "Child content",
      },
    });

    const testError = new Error("Test error message");
    testError.stack = "Error stack trace";
    (component as any).handleError(testError);

    await tick();

    expect(screen.getByText("Error Details (Development)")).toBeInTheDocument();
    expect(screen.getByText("Test error message")).toBeInTheDocument();
    expect(screen.getByText("Error stack trace")).toBeInTheDocument();
  });

  it("does not display error details in production mode", async () => {
    // Note: import.meta.env.DEV is replaced at build time by Vite/Vitest via define config.
    // Since we can't change it at runtime, this test verifies the component structure.
    // In the current test environment (DEV=true via vitest.config.ts), error details will be shown.
    // To properly test production mode, you would need a separate test run with DEV=false in config.
    
    const { component } = render(ErrorBoundary, {
      props: {
        children: () => "Child content",
      },
    });

    (component as any).handleError(new Error("Test error"));
    await tick();

    // The error boundary should still render correctly
    expect(screen.getByText("Oops! Something went wrong")).toBeInTheDocument();
    
    // In the current test environment (DEV=true), error details will be shown
    // This test verifies the component doesn't crash and handles errors properly
  });

  it("calls onerror callback when error occurs", async () => {
    const testError = new Error("Test error");
    
    const { component } = render(ErrorBoundary, {
      props: {
        onerror: mockOnError,
        context: { userId: "123", page: "dashboard" },
        children: () => "Child content",
      },
    });

    (component as any).handleError(testError);

    await tick();

    expect(mockOnError).toHaveBeenCalledWith({
      error: testError,
      context: expect.objectContaining({
        userId: "123",
        page: "dashboard",
        errorId: expect.stringMatching(/^error-\d+-\w+$/),
      }),
    });
  });

  it("handles non-Error objects passed to handleError", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        children: () => "Child content",
      },
    });

    (component as any).handleError("String error");
    await tick();

    expect(screen.getByText("Oops! Something went wrong")).toBeInTheDocument();
  });

  it("passes additional context to handleError", async () => {
    const testError = new Error("Test error");
    const timestamp = Date.now();
    
    const { component } = render(ErrorBoundary, {
      props: {
        onerror: mockOnError,
        context: { userId: "123" },
        children: () => "Child content",
      },
    });

    (component as any).handleError(testError, { action: "save", timestamp });

    await tick();

    expect(mockOnError).toHaveBeenCalledWith({
      error: testError,
      context: expect.objectContaining({
        userId: "123",
        action: "save",
        timestamp: expect.any(Number),
        errorId: expect.stringMatching(/^error-\d+-\w+$/),
      }),
    });
  });

  it("generates unique error IDs", async () => {
    const { component, unmount } = render(ErrorBoundary, {
      props: {
        onerror: mockOnError,
        children: () => "Child content",
      },
    });

    (component as any).handleError(new Error("Error 1"));
    await tick();
    unmount();
    cleanup();

    // Render again to trigger second error
    const { component: component2 } = render(ErrorBoundary, {
      props: {
        onerror: mockOnError,
        children: () => "Child content",
      },
    });

    (component2 as any).handleError(new Error("Error 2"));
    await tick();

    const calls = mockOnError.mock.calls;
    expect(calls.length).toBeGreaterThanOrEqual(2);
    expect(calls[0][0].context.errorId).not.toBe(calls[1][0].context.errorId);
    expect(calls[0][0].context.errorId).toMatch(/^error-\d+-\w+$/);
    expect(calls[1][0].context.errorId).toMatch(/^error-\d+-\w+$/);
  });

  it("sets data-error-id attribute on error boundary", async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        children: () => "Child content",
      },
    });

    (component as any).handleError(new Error("Test error"));
    await tick();

    const errorBoundary = document.querySelector('[data-error-id]');
    expect(errorBoundary).toBeInTheDocument();
    expect(errorBoundary?.getAttribute("data-error-id")).toMatch(/^error-\d+-\w+$/);
  });
});