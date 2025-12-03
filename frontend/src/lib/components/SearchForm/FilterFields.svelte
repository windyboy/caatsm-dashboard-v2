<!--
  @component FilterFields
  Filter input fields for search form.
  
  @param {string} type - Message type filter value
  @param {string} priority - Priority filter value
  @param {string} start_time - Start time filter value
  @param {string} end_time - End time filter value
  @param {(field: string, value: string) => void} onchange - Called when filter value changes
-->

<script lang="ts">
  interface Props {
    type: string;
    priority: string;
    start_time: string;
    end_time: string;
    onchange: (field: string, value: string) => void;
  }

  let { type, priority, start_time, end_time, onchange }: Props = $props();
</script>

<div class="filters-grid">
  <div class="filters-header">
    <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
      ></path>
    </svg>
    <span class="text-sm font-semibold text-brand-700">Filters</span>
  </div>
  <div></div>
  <div class="filter-field">
    <label for="type-select" class="filter-label">Type</label>
    <select
      id="type-select"
      value={type}
      onchange={(e) => onchange("type", e.currentTarget.value)}
      class="filter-select"
    >
      <option value="">Any</option>
      <option value="AFTN">AFTN</option>
      <option value="SITA">SITA</option>
      <option value="ACARS">ACARS</option>
      <option value="CPDLC">CPDLC</option>
    </select>
  </div>
  <div class="filter-field">
    <label for="priority-select" class="filter-label">Priority</label>
    <select
      id="priority-select"
      value={priority}
      onchange={(e) => onchange("priority", e.currentTarget.value)}
      class="filter-select"
    >
      <option value="">Any</option>
      <option value="1">Urgent</option>
      <option value="2">Operational</option>
      <option value="3">Routine</option>
    </select>
  </div>
  <div class="filter-field">
    <label for="start-time-input" class="filter-label">Start Time</label>
    <input
      id="start-time-input"
      type="datetime-local"
      value={start_time}
      onchange={(e) => onchange("start_time", e.currentTarget.value)}
      class="filter-input"
    />
  </div>
  <div class="filter-field">
    <label for="end-time-input" class="filter-label">End Time</label>
    <input
      id="end-time-input"
      type="datetime-local"
      value={end_time}
      onchange={(e) => onchange("end_time", e.currentTarget.value)}
      class="filter-input"
    />
  </div>
</div>

<style>
  .filters-grid {
    display: grid;
    gap: 1.25rem;
  }

  @media (min-width: 768px) {
    .filters-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  .filters-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .filter-label {
    display: block;
    font-size: 0.75rem;
    line-height: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    margin-bottom: 0.625rem;
    font-weight: 700;
    color: #475569;
  }

  .filter-select,
  .filter-input {
    width: 100%;
    border-radius: 0.5rem;
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    border: 0;
    background: linear-gradient(to right, #f8fafc, rgba(241, 245, 249, 0.5));
    color: #0f172a;
    transition: all 0.3s ease;
    box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.08);
  }

  .filter-select:hover,
  .filter-input:hover {
    box-shadow:
      0 4px 6px -1px rgba(15, 23, 42, 0.1),
      0 2px 4px -2px rgba(15, 23, 42, 0.1);
  }

  .filter-select:focus,
  .filter-input:focus {
    background: rgba(255, 255, 255, 0.95);
    outline: none;
    box-shadow: var(--input-focus-shadow);
  }

  .filter-input::placeholder {
    color: #94a3b8;
  }
</style>

