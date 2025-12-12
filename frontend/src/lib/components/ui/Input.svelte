<script lang="ts">
  import { Input as ShadcnInput } from "./input/index.js";
  import { Label } from "./label/index.js";
  import type { HTMLInputAttributes } from "svelte/elements";

  let inputCounter = 0;

  /**
   * Input component wrapper that combines shadcn Input with optional Label.
   * Use this component when you need an input field with a label.
   * For inputs without labels, use the base Input component directly from "./input/index.js".
   *
   * @example
   * ```svelte
   * <Input
   *   label="Email"
   *   name="email"
   *   type="email"
   *   placeholder="Enter your email"
   *   bind:value={email}
   * />
   * ```
   */
  interface InputProps extends Omit<
    HTMLInputAttributes,
    | "value"
    | "type"
    | "name"
    | "disabled"
    | "class"
    | "id"
    | "placeholder"
    | "aria-label"
    | "aria-describedby"
  > {
    /** Optional label text displayed above the input */
    label?: string;
    /** Input value (bindable) */
    value?: string;
    /** Input type (default: "text") */
    type?: Exclude<HTMLInputAttributes["type"], null>;
    /** Input name attribute */
    name?: string;
    /** Whether the input is disabled */
    disabled?: boolean;
    /** Aria label for accessibility (falls back to label if not provided) */
    ariaLabel?: string;
    /** Aria describedby for linking error messages or descriptions */
    "aria-describedby"?: string;
    /** Additional CSS classes */
    class?: string;
    /** Input placeholder */
    placeholder?: string;
    /** Input ID */
    id?: string;
  }

  let {
    label,
    value = $bindable(""),
    placeholder = "",
    type = "text",
    name = "",
    disabled = false,
    ariaLabel,
    "aria-describedby": ariaDescribedBy,
    class: className,
    id,
    ...restProps
  }: InputProps = $props();

  // Type for restProps - all remaining HTMLInputAttributes that weren't destructured
  // Exclude files since we're not using file type
  type RestProps = Omit<
    HTMLInputAttributes,
    | "value"
    | "type"
    | "name"
    | "disabled"
    | "class"
    | "id"
    | "placeholder"
    | "aria-label"
    | "aria-describedby"
    | "label"
    | "ariaLabel"
    | "files"
  > & {
    files?: never;
  };

  // Generate unique ID for label-input association if not provided
  // Use name if available, otherwise generate a unique ID
  const inputId = $derived(
    id || (label && name ? name : label ? `input-${++inputCounter}` : undefined)
  );

  // Ensure type is never null for the underlying component
  const inputType = $derived(type ?? "text");
</script>

<div class="flex flex-col gap-1.5">
  {#if label}
    <Label for={inputId} class="text-muted-foreground text-sm">
      {label}
    </Label>
  {/if}
  <ShadcnInput
    id={inputId}
    type={inputType as Exclude}
    {name}
    {placeholder}
    bind:value
    {disabled}
    aria-label={ariaLabel || (label && !inputId ? label : undefined)}
    aria-describedby={ariaDescribedBy}
    class={className}
    {...restProps as RestProps}
  />
</div>
