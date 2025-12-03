<!--
  @component TimeRangeSelector
  Provides time range selection with presets (1h, 24h, 7d, 30d, 90d) and custom range.
  
  Features:
  - Preset buttons for quick selection
  - Custom date/time picker for precise ranges
  - Updates timeRange store on change
-->

<script lang="ts">
  import { timeRange, type TimeRangePreset } from "$lib/stores/data/timeRange";
  import { createLogger } from "$lib/utils/logger";

  const logger = createLogger("TimeRangeSelector");

  // Use derived to get current range from store (reactive, no loop)
  const currentRange = $derived($timeRange);
  
  // Local state for custom picker inputs (only for editing, not synced back)
  let customStart = $state(new Date($timeRange.start));
  let customEnd = $state(new Date($timeRange.end));
  
  // Show custom picker based on current preset
  const showCustomPicker = $derived(currentRange.preset === "custom");

  // Sync custom dates when store changes (only if preset is custom and dates actually changed)
  $effect(() => {
    if (currentRange.preset === "custom") {
      const newStart = new Date(currentRange.start);
      const newEnd = new Date(currentRange.end);
      // Only update if dates actually changed (avoid unnecessary updates)
      if (customStart.getTime() !== newStart.getTime()) {
        customStart = newStart;
      }
      if (customEnd.getTime() !== newEnd.getTime()) {
        customEnd = newEnd;
      }
    }
  });

  const presets: Array<{ value: TimeRangePreset; label: string }> = [
    { value: "1h", label: "1 Hour" },
    { value: "24h", label: "24 Hours" },
    { value: "7d", label: "7 Days" },
    { value: "30d", label: "30 Days" },
    { value: "90d", label: "90 Days" },
    { value: "custom", label: "Custom" },
  ];

  function handlePresetSelect(preset: TimeRangePreset) {
    if (preset === "custom") {
      // If already on custom, keep current dates; otherwise initialize from current range
      if (currentRange.preset !== "custom") {
        customStart = new Date(currentRange.start);
        customEnd = new Date(currentRange.end);
      }
      // setCustomRange already sets preset to "custom", so we don't need setPreset
      timeRange.setCustomRange(customStart, customEnd);
      logger.debug("Time range preset selected", { preset });
    } else {
      timeRange.setPreset(preset);
      logger.debug("Time range preset selected", { preset });
    }
  }

  function handleCustomRangeApply() {
    // Validate dates
    if (customStart >= customEnd) {
      logger.warn("Invalid time range: start must be before end");
      return;
    }

    // Ensure end is not in the future
    const now = new Date();
    if (customEnd > now) {
      customEnd = new Date(now);
    }

    timeRange.setCustomRange(customStart, customEnd);
    logger.debug("Custom time range applied", {
      start: customStart.toISOString(),
      end: customEnd.toISOString(),
    });
  }

  function formatDate(date: Date): string {
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  function formatDateTime(date: Date): string {
    return date.toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }
</script>

<div class="time-range-selector" role="group" aria-label="Time Range Selection">
  <div class="selector-header">
    <p class="selector-title">Time Range</p>
  </div>

  <div class="preset-buttons">
    {#each presets as preset}
      <button
        type="button"
        class="preset-button"
        class:active={currentRange.preset === preset.value}
        onclick={() => handlePresetSelect(preset.value)}
        aria-pressed={currentRange.preset === preset.value}
      >
        {preset.label}
      </button>
    {/each}
  </div>

  {#if showCustomPicker}
    <div class="custom-picker">
      <div class="picker-row">
        <label for="custom-start" class="picker-label">Start:</label>
        <input
          id="custom-start"
          type="datetime-local"
          bind:value={customStart}
          class="picker-input"
          max={customEnd.toISOString().slice(0, 16)}
        />
      </div>
      <div class="picker-row">
        <label for="custom-end" class="picker-label">End:</label>
        <input
          id="custom-end"
          type="datetime-local"
          bind:value={customEnd}
          class="picker-input"
          max={new Date().toISOString().slice(0, 16)}
        />
      </div>
      <button
        type="button"
        class="apply-button"
        onclick={handleCustomRangeApply}
        disabled={customStart >= customEnd}
      >
        Apply
      </button>
    </div>
  {/if}

  <div class="current-range">
    <p class="range-text">
      {formatDateTime(currentRange.start)} - {formatDateTime(currentRange.end)}
    </p>
  </div>
</div>

<style>
  .time-range-selector {
    border-radius: 0.5rem;
    padding: 1rem;
    background: var(--stats-card-bg);
    box-shadow: var(--stats-card-shadow);
    border: var(--stats-card-border);
  }

  .selector-header {
    margin-bottom: 0.75rem;
  }

  .selector-title {
    font-size: 0.75rem;
    line-height: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 700;
    background: linear-gradient(to right, #2563eb, #9333ea);
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .preset-buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .preset-button {
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    border: 1px solid rgba(148, 163, 184, 0.3);
    background: rgba(248, 250, 252, 0.8);
    color: #334155;
    cursor: pointer;
    transition:
      background-color 0.2s,
      border-color 0.2s,
      transform 0.2s;
  }

  .preset-button:hover {
    background: rgba(241, 245, 249, 1);
    border-color: rgba(148, 163, 184, 0.5);
    transform: translateY(-1px);
  }

  .preset-button.active {
    background: linear-gradient(to right, #2563eb, #9333ea);
    color: white;
    border-color: transparent;
    box-shadow: 0 2px 4px rgba(37, 99, 235, 0.2);
  }

  .custom-picker {
    padding: 1rem;
    border-radius: 0.375rem;
    background: rgba(248, 250, 252, 0.5);
    margin-bottom: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .picker-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .picker-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #334155;
    min-width: 4rem;
  }

  .picker-input {
    flex: 1;
    padding: 0.5rem;
    border-radius: 0.375rem;
    border: 1px solid rgba(148, 163, 184, 0.3);
    background: white;
    font-size: 0.875rem;
    color: #334155;
  }

  .picker-input:focus {
    outline: none;
    border-color: #2563eb;
    box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
  }

  .apply-button {
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 600;
    border: none;
    background: linear-gradient(to right, #2563eb, #9333ea);
    color: white;
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;
  }

  .apply-button:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 6px rgba(37, 99, 235, 0.2);
  }

  .apply-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .current-range {
    padding-top: 0.75rem;
    border-top: 1px solid rgba(148, 163, 184, 0.2);
  }

  .range-text {
    font-size: 0.75rem;
    color: #64748b;
    text-align: center;
  }
</style>

