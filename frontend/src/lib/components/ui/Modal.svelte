<!--
  @component Modal
  Standardized modal/dialog component with consistent API following design system.

  @param {boolean} open - Whether the modal is open
  @param {'sm' | 'md' | 'lg' | 'xl' | 'full'} size - Modal size
  @param {boolean} closable - Whether to show close button
  @param {boolean} backdropClosable - Whether clicking backdrop closes modal
  @param {string} title - Modal title
  @param {boolean} footer - Whether modal has footer content
  @param {boolean} centered - Whether modal is centered
  @param {string | number} width - Custom width
-->

<script lang="ts">
  import type { Snippet } from "svelte";
  import { onDestroy, tick } from "svelte";
  import type { ModalProps } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends ModalProps {
    onopen?: (event: Event) => void;
    onclose?: (event: Event) => void;
    onbackdropClick?: (event: MouseEvent) => void;
    children?: Snippet;
    footerSnippet?: Snippet;
  }

  let {
    open = $bindable(false),
    size = 'md',
    closable = true,
    backdropClosable = true,
    title,
    footer = false,
    centered = true,
    width,
    disabled = false,
    class: className = '',
    onopen,
    onclose,
    onbackdropClick,
    children,
    footerSnippet
  }: Props = $props();

  // Internal state
  let modalElement = $state<HTMLElement>();
  let previousFocus = $state<HTMLElement>();
  let focusableElements = $state<HTMLElement[]>([]);
  let firstFocusable = $state<HTMLElement>();
  let lastFocusable = $state<HTMLElement>();

  // Size classes using design tokens
  const sizeClasses = {
    sm: "max-w-sm",
    md: "max-w-md",
    lg: "max-w-lg",
    xl: "max-w-xl",
    full: "max-w-full"
  };

  // Computed values with $derived
  const positionClasses = $derived(centered
    ? "items-center justify-center"
    : "items-start justify-center pt-16");
  const customWidth = $derived(width ? `width: ${typeof width === 'number' ? `${width}px` : width};` : '');

  // Get all focusable elements within the modal
  function getFocusableElements(): HTMLElement[] {
    if (!modalElement) return [];

    const focusableSelectors = [
      'a[href]',
      'button:not([disabled])',
      'textarea:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      '[tabindex]:not([tabindex="-1"])'
    ].join(', ');

    const elements = Array.from(modalElement.querySelectorAll<HTMLElement>(focusableSelectors));
    return elements.filter(el => {
      // Filter out hidden or invisible elements
      return el.offsetParent !== null && !el.hasAttribute('hidden');
    });
  }

  // Setup focus trap
  function setupFocusTrap() {
    focusableElements = getFocusableElements();
    firstFocusable = focusableElements[0];
    lastFocusable = focusableElements[focusableElements.length - 1];

    // Focus the first focusable element or modal container
    if (firstFocusable) {
      firstFocusable.focus();
    } else if (modalElement) {
      modalElement.focus();
    }
  }

  // Handle Tab key to trap focus
  function handleFocusTrap(event: KeyboardEvent) {
    if (event.key !== 'Tab') return;
    if (!firstFocusable || !lastFocusable) return;

    // Shift+Tab: if on first element, jump to last
    if (event.shiftKey) {
      if (document.activeElement === firstFocusable) {
        event.preventDefault();
        lastFocusable.focus();
      }
    } else {
      // Tab: if on last element, jump to first
      if (document.activeElement === lastFocusable) {
        event.preventDefault();
        firstFocusable.focus();
      }
    }
  }

  async function handleOpen() {
    // Store previous focus
    previousFocus = document.activeElement as HTMLElement;

    // Setup focus trap after modal is rendered
    await tick();
    setupFocusTrap();

    onopen?.(new Event('open'));
  }

  function handleClose() {
    open = false;
    onclose?.(new Event('close'));

    // Restore previous focus
    if (previousFocus) {
      previousFocus.focus();
    }
  }

  function handleBackdropClick(event: Event) {
    if (event.target === event.currentTarget && backdropClosable) {
      // Only invoke the callback if it's actually a MouseEvent
      if (onbackdropClick && event instanceof MouseEvent) {
        onbackdropClick(event);
      }
      handleClose();
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && closable) {
      handleClose();
    }
  }

  // Handle open/close effects with $effect
  $effect(() => {
    if (open) {
      handleOpen();
      document.body.classList.add('modal-open');
    } else {
      document.body.classList.remove('modal-open');
    }
  });

  // Cleanup on component destroy
  onDestroy(() => {
    document.body.classList.remove('modal-open');
  });
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <!-- Daisy UI Modal -->
  <div 
    class="modal modal-open" 
    role="presentation"
    onclick={handleBackdropClick}
    onkeydown={(e) => e.key === 'Enter' && handleBackdropClick(e)}
  >
    <div 
      bind:this={modalElement}
      class="modal-box {sizeClasses[size!]} {className}" 
      style={customWidth}
      tabindex="-1"
      role="dialog"
      aria-modal="true"
      onkeydown={handleFocusTrap}
    >
      <!-- Header -->
      {#if title || closable}
        <div class="flex items-center justify-between mb-4">
          {#if title}
            <h3 class="font-bold text-lg">{title}</h3>
          {/if}

          {#if closable}
            <button
              type="button"
              class="btn btn-sm btn-circle btn-ghost"
              onclick={handleClose}
              aria-label="Close modal"
            >
              ✕
            </button>
          {/if}
        </div>
      {/if}

      <!-- Body -->
      {@render children?.()}

      <!-- Footer -->
      {#if footer && footerSnippet}
        <div class="modal-action">
          {@render footerSnippet()}
        </div>
      {/if}
    </div>
  </div>
{/if}
