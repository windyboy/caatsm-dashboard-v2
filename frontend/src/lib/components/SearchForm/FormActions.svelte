<!--
  @component FormActions
  Form action buttons (reset and submit).
  
  @param {string | null} error - Error message to display
  @param {() => void} onreset - Called when reset button is clicked
-->

<script lang="ts">
  interface Props {
    error?: string | null;
    onreset?: () => void;
  }

  let { error, onreset }: Props = $props();
</script>

{#if error}
  <div class="error-message">
    <div class="error-icon">⚠️</div>
    <span class="error-text">{error}</span>
  </div>
{/if}
<div class="form-actions">
  <button type="button" class="btn-reset" onclick={(e) => {
    e.preventDefault();
    onreset?.();
  }}>
    Reset
  </button>
  <button type="submit" class="btn-search">Search</button>
</div>

<style>
  .error-message {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
    border-radius: 0.5rem;
    border: 1px solid rgba(254, 202, 202, 0.6);
    background: linear-gradient(to right, #fef2f2, #fff7ed);
    color: #b91c1c;
  }

  .error-icon {
    flex-shrink: 0;
    font-size: 1.125rem;
  }

  .error-text {
    font-size: 0.875rem;
    line-height: 1.25rem;
    font-weight: 600;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    padding-top: 1rem;
    border-top: 1px solid rgba(226, 232, 240, 0.6);
  }

  .btn-reset {
    border-radius: 0.5rem;
    padding: 0.625rem 1.5rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    font-weight: 600;
    border: 0;
    color: #334155;
    background: linear-gradient(to right, rgba(241, 245, 249, 0.8), rgba(226, 232, 240, 0.6));
    box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.06);
    transition: all 0.3s ease;
    cursor: pointer;
  }

  .btn-reset:hover {
    background: linear-gradient(to right, rgba(226, 232, 240, 0.9), rgba(203, 213, 225, 0.7));
    box-shadow:
      0 4px 6px -1px rgba(15, 23, 42, 0.1),
      0 2px 4px -2px rgba(15, 23, 42, 0.1);
    transform: var(--btn-hover-lift);
  }

  .btn-search {
    border-radius: 0.5rem;
    padding: 0.625rem 1.5rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    font-weight: 600;
    color: #ffffff;
    border: 0;
    background: linear-gradient(to right, #3b82f6, #a855f7);
    box-shadow:
      0 10px 15px -3px rgba(15, 23, 42, 0.2),
      0 4px 6px -4px rgba(15, 23, 42, 0.1);
    transition: all 0.3s ease;
    cursor: pointer;
    outline: none;
  }

  .btn-search:hover {
    background: linear-gradient(to right, #2563eb, #9333ea);
    box-shadow:
      0 12px 18px -3px rgba(15, 23, 42, 0.25),
      0 4px 6px -4px rgba(15, 23, 42, 0.15);
    transform: var(--btn-hover-lift);
  }

  .btn-search:focus {
    outline: none;
    box-shadow:
      0 0 0 2px rgba(59, 130, 246, 0.4),
      var(--shadow-xl);
  }
</style>

