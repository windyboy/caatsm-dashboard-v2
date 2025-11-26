<!--
  @component LoadingSpinner
  Animated loading spinner with multiple variants for different use cases.

  @param {string} size - Size variant: 'sm', 'md', 'lg', 'xl'
  @param {string} variant - Visual variant: 'primary', 'secondary', 'gradient'
  @param {string} message - Optional loading message
-->

<script lang="ts">
  /** @type {'sm' | 'md' | 'lg' | 'xl'} */
  export let size: 'sm' | 'md' | 'lg' | 'xl' = 'md';
  /** @type {'primary' | 'secondary' | 'gradient'} */
  export let variant: 'primary' | 'secondary' | 'gradient' = 'primary';
  /** @type {string} */
  export let message: string = '';

  const sizeClasses = {
    sm: 'w-4 h-4',
    md: 'w-8 h-8',
    lg: 'w-12 h-12',
    xl: 'w-16 h-16'
  };

  const variantClasses = {
    primary: 'text-brand-500',
    secondary: 'text-slate-400',
    gradient: 'text-transparent bg-gradient-to-r from-brand-500 via-accent-500 to-success-500 bg-clip-text'
  };
</script>

<div class="flex flex-col items-center justify-center gap-3 p-4">
  <div class="relative {sizeClasses[size]}">
    <!-- Outer ring -->
    <div class="absolute inset-0 rounded-full border-2 border-slate-200"></div>

    <!-- Spinning ring -->
    <div class="absolute inset-0 rounded-full border-2 border-t-transparent {variantClasses[variant]} animate-spin"></div>

    <!-- Inner pulse -->
    {#if variant === 'gradient'}
      <div class="absolute inset-1 rounded-full bg-gradient-to-r from-brand-400 via-accent-400 to-success-400 animate-pulse opacity-30"></div>
    {/if}
  </div>

  {#if message}
    <p class="text-sm text-slate-600 font-medium animate-pulse">{message}</p>
  {/if}
</div>

<style>
  .animate-spin {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>