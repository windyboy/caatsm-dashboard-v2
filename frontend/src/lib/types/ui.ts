// Design System Types and Interfaces

// ============================================================================
// COMMON TYPES
// ============================================================================

/**
 * Standard size variants used across all components
 */
export type ComponentSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

/**
 * Standard component variants for different visual styles
 */
export type ComponentVariant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'success' | 'warning' | 'danger';

/**
 * Common component props that all components should support
 */
export interface BaseComponentProps {
  /** Additional CSS classes to apply */
  class?: string;
  /** Whether the component is disabled */
  disabled?: boolean;
  /** Whether the component is in a loading state */
  loading?: boolean;
  /** Test ID for testing purposes */
  'data-testid'?: string;
}

/**
 * Event handler types for common interactions
 */
export interface ComponentEvents<T = Event> {
  click?: T;
  focus?: T;
  blur?: T;
  mouseenter?: T;
  mouseleave?: T;
}

// ============================================================================
// BUTTON COMPONENT
// ============================================================================

export interface ButtonProps extends BaseComponentProps {
  /** Button variant */
  variant?: ComponentVariant;
  /** Button size */
  size?: ComponentSize;
  /** Button type attribute */
  type?: 'button' | 'submit' | 'reset';
  /** If provided, renders as anchor tag */
  href?: string;
  /** Whether to enable ripple effect */
  ripple?: boolean;
  /** Whether the button should take full width */
  fullWidth?: boolean;
  /** Click event callback */
  onclick?: (event: Event) => void;
}

// ============================================================================
// CARD COMPONENT
// ============================================================================

export interface CardProps extends BaseComponentProps {
  /** Card variant */
  variant?: 'default' | 'elevated' | 'bordered' | 'filled';
  /** Card size affecting padding */
  size?: ComponentSize;
  /** Whether to enable hover effects */
  hover?: boolean;
  /** Whether to apply padding */
  padding?: boolean;
  /** Whether the card is interactive/clickable */
  interactive?: boolean;
  /** Click event callback */
  onclick?: (event: Event) => void;
}

// ============================================================================
// INPUT COMPONENT
// ============================================================================

export interface InputProps extends BaseComponentProps {
  /** Input type */
  type?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url' | 'search';
  /** Input variant */
  variant?: ComponentVariant;
  /** Input size */
  size?: ComponentSize;
  /** Placeholder text */
  placeholder?: string;
  /** Input value */
  value?: string | number;
  /** Whether the input is required */
  required?: boolean;
  /** Whether the input is readonly */
  readonly?: boolean;
  /** Minimum length for validation */
  minLength?: number;
  /** Maximum length for validation */
  maxLength?: number;
  /** Pattern for validation */
  pattern?: string;
  /** Error message to display */
  error?: string;
  /** Helper text to display */
  helperText?: string;
  /** Leading icon (slot-based) */
  leadingIcon?: boolean;
  /** Trailing icon (slot-based) */
  trailingIcon?: boolean;
  /** Whether to show character count */
  showCount?: boolean;
  /** Input event callback */
  oninput?: (event: Event) => void;
  /** Change event callback */
  onchange?: (event: Event) => void;
  /** Focus event callback */
  onfocus?: (event: Event) => void;
  /** Blur event callback */
  onblur?: (event: Event) => void;
}

// ============================================================================
// SELECT COMPONENT
// ============================================================================

export interface SelectOption {
  /** Option value */
  value: string | number;
  /** Option label */
  label: string;
  /** Whether option is disabled */
  disabled?: boolean;
  /** Option group */
  group?: string;
}

export interface SelectProps extends BaseComponentProps {
  /** Select variant */
  variant?: ComponentVariant;
  /** Select size */
  size?: ComponentSize;
  /** Available options */
  options: SelectOption[];
  /** Selected value */
  value?: string | number | (string | number)[];
  /** Placeholder text */
  placeholder?: string;
  /** Whether multiple selection is allowed */
  multiple?: boolean;
  /** Whether the select is searchable */
  searchable?: boolean;
  /** Whether the select is clearable */
  clearable?: boolean;
  /** Maximum number of options to show */
  maxOptions?: number;
  /** Error message to display */
  error?: string;
  /** Helper text to display */
  helperText?: string;
  /** Loading state text */
  loadingText?: string;
  /** Change event callback */
  onchange?: (detail: { value: string | number | (string | number)[] }) => void;
  /** Open event callback */
  onopen?: (event: Event) => void;
  /** Close event callback */
  onclose?: (event: Event) => void;
}

// ============================================================================
// MODAL COMPONENT
// ============================================================================

export interface ModalProps extends BaseComponentProps {
  /** Whether the modal is open */
  open: boolean;
  /** Modal size */
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  /** Whether to show close button */
  closable?: boolean;
  /** Whether clicking backdrop closes modal */
  backdropClosable?: boolean;
  /** Modal title */
  title?: string;
  /** Modal footer content */
  footer?: boolean;
  /** Whether modal is centered */
  centered?: boolean;
  /** Custom width */
  width?: string | number;
  /** Open event callback */
  onopen?: (event: Event) => void;
  /** Close event callback */
  onclose?: (event: Event) => void;
  /** Backdrop click event callback */
  onbackdropClick?: (event: MouseEvent) => void;
}

// ============================================================================
// TOAST COMPONENT
// ============================================================================

export interface ToastProps extends BaseComponentProps {
  /** Toast variant */
  variant?: ComponentVariant;
  /** Toast title */
  title?: string;
  /** Toast message */
  message?: string;
  /** Toast duration in milliseconds */
  duration?: number;
  /** Whether toast can be dismissed */
  dismissible?: boolean;
  /** Custom icon */
  icon?: string;
  /** Toast position */
  position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'top-center' | 'bottom-center';
  /** Dismiss event callback */
  ondismiss?: (event: Event) => void;
  /** Timeout event callback */
  ontimeout?: (event: Event) => void;
}

// ============================================================================
// LOADING SPINNER COMPONENT
// ============================================================================

export interface LoadingSpinnerProps extends BaseComponentProps {
  /** Spinner size */
  size?: ComponentSize;
  /** Spinner variant */
  variant?: ComponentVariant;
  /** Loading message */
  message?: string;
  /** Whether to show overlay */
  overlay?: boolean;
  /** Overlay color */
  overlayColor?: string;
}

// ============================================================================
// THEME SYSTEM
// ============================================================================

/**
 * Design tokens for the theme system
 */
export interface DesignTokens {
  // Colors
  colors: {
    brand: Record<string, string>;
    accent: Record<string, string>;
    success: Record<string, string>;
    warning: Record<string, string>;
    danger: Record<string, string>;
    neutral: Record<string, string>;
  };

  // Spacing scale
  spacing: Record<string, string>;

  // Typography
  typography: {
    fontFamily: Record<string, string>;
    fontSize: Record<string, string>;
    fontWeight: Record<string, string>;
    lineHeight: Record<string, string>;
    letterSpacing: Record<string, string>;
  };

  // Border radius
  borderRadius: Record<string, string>;

  // Shadows
  shadows: Record<string, string>;

  // Transitions
  transitions: {
    duration: Record<string, string>;
    easing: Record<string, string>;
  };
}

// ============================================================================
// COMPOSITION PATTERNS
// ============================================================================

/**
 * Props for components that support loading states
 */
export interface WithLoadingProps {
  loadingText?: string;
}

/**
 * Props for components that support error states
 */
export interface WithErrorProps {
  error?: string | Error;
  showError?: boolean;
  /** Retry event callback */
  onretry?: (event: Event) => void;
}

/**
 * Props for components that support empty states
 */
export interface WithEmptyProps {
  empty?: boolean;
  emptyText?: string;
  emptyIcon?: string;
}

/**
 * Combined props for components with common behaviors
 */
export interface EnhancedComponentProps extends
  BaseComponentProps,
  WithLoadingProps,
  WithErrorProps,
  WithEmptyProps {}

// ============================================================================
// COMPONENT SLOTS
// ============================================================================

/**
 * Standard slot interfaces for component composition
 */
export interface ComponentSlots {
  default?: any;
  header?: any;
  footer?: any;
  leading?: any;
  trailing?: any;
  actions?: any;
  content?: any;
}