<!--
  @component Card
  Reusable card wrapper component with glassmorphism effects.

  @param {string} variant - Card style variant: 'default', 'elevated', 'bordered'
  @param {boolean} hover - Whether to enable hover effects
  @param {boolean} padding - Whether to include default padding
  @param {string} className - Additional CSS classes
-->

<script lang="ts">
  /** @type {'default' | 'elevated' | 'bordered'} */
  export let variant: 'default' | 'elevated' | 'bordered' = 'default';
  /** @type {boolean} */
  export let hover: boolean = true;
  /** @type {boolean} */
  export let padding: boolean = true;
  /** @type {string} */
  export let className: string = '';

  const baseClasses = "rounded-lg backdrop-blur-md border-0 transition-all duration-300";

  const variantClasses = {
    default: "bg-white/95 card-glow",
    elevated: "bg-white/98 shadow-xl card-glow",
    bordered: "bg-white/90 border-2 border-brand-200/50"
  };

  const paddingClasses = padding ? "p-4 sm:p-6 lg:p-8" : "";

  const hoverClasses = hover ? "card-hover-pulse" : "";

  $: classes = `${baseClasses} ${variantClasses[variant]} ${paddingClasses} ${hoverClasses} ${className}`;
</script>

<div class={classes}>
  <slot />
</div>