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
  import { fade } from "svelte/transition";
  import Button from "./Button.svelte";
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

  function handleOpen() {
    // Store previous focus
    previousFocus = document.activeElement as HTMLElement;

    // Focus modal
    setTimeout(() => {
      modalElement?.focus();
    }, 100);

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

  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget && backdropClosable) {
      onbackdropClick?.(event);
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
    }
  });
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <!-- Backdrop -->
  <div
    class="modal-backdrop"
    transition:fade={{ duration: 200 }}
    onclick={handleBackdropClick}
  >
    <!-- Modal container -->
    <div
      class="modal-container {positionClasses}"
      transition:fade={{ duration: 200, delay: 50 }}
    >
      <!-- Modal content -->
      <div
        bind:this={modalElement}
        class="modal-content {sizeClasses[size!]} {className}"
        style={customWidth}
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? "modal-title" : undefined}
        tabindex="-1"
      >
        <!-- Header -->
        {#if title || closable}
          <div class="modal-header">
            {#if title}
              <h2 id="modal-title" class="modal-title">{title}</h2>
            {/if}

            {#if closable}
              <button
                type="button"
                class="modal-close"
                onclick={handleClose}
                aria-label="Close modal"
              >
                ×
              </button>
            {/if}
          </div>
        {/if}

        <!-- Body -->
        <div class="modal-body">
          {@render children?.()}
        </div>

        <!-- Footer -->
        {#if footer && footerSnippet}
          <div class="modal-footer">
            {@render footerSnippet()}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    z-index: 1000;
    display: flex;
    padding: var(--spacing-lg);
  }

  .modal-container {
    display: flex;
    width: 100%;
    height: 100%;
    max-height: 100vh;
    overflow-y: auto;
  }

  .modal-content {
    background: white;
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-2xl);
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - 2 * var(--spacing-lg));
    margin: auto;
    position: relative;
    transform: scale(0.95);
    transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .modal-backdrop .modal-content {
    transform: scale(1);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--spacing-lg);
    border-bottom: 1px solid var(--color-slate-100);
    flex-shrink: 0;
  }

  .modal-title {
    font-size: var(--text-xl);
    font-weight: var(--font-semibold);
    color: var(--color-slate-900);
    margin: 0;
  }

  .modal-close {
    background: none;
    border: none;
    color: var(--color-slate-400);
    cursor: pointer;
    font-size: 1.5rem;
    line-height: 1;
    padding: var(--spacing-xs);
    border-radius: var(--radius-md);
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
  }

  .modal-close:hover {
    background-color: var(--color-slate-100);
    color: var(--color-slate-600);
  }

  .modal-body {
    padding: var(--spacing-lg);
    flex: 1;
    overflow-y: auto;
    color: var(--color-slate-700);
  }

  .modal-footer {
    padding: var(--spacing-lg);
    border-top: 1px solid var(--color-slate-100);
    flex-shrink: 0;
    display: flex;
    gap: var(--spacing-sm);
    justify-content: flex-end;
  }

  /* Responsive adjustments */
  @media (max-width: 640px) {
    .modal-backdrop {
      padding: var(--spacing-md);
    }

    .modal-content {
      max-height: calc(100vh - 2 * var(--spacing-md));
    }

    .modal-header,
    .modal-body,
    .modal-footer {
      padding: var(--spacing-md);
    }
  }

  /* Focus trap styles */
  .modal-content:focus {
    outline: none;
  }

  /* Prevent body scroll when modal is open */
  :global(body.modal-open) {
    overflow: hidden;
  }
</style>