export type ComponentSize = "xs" | "sm" | "md" | "lg" | "xl";

export type ComponentVariant =
  | "primary"
  | "secondary"
  | "outline"
  | "ghost"
  | "success"
  | "warning"
  | "danger";

export interface BaseComponentProps {
  class?: string;
  disabled?: boolean;
  loading?: boolean;
  "data-testid"?: string;
}

export interface ButtonProps extends BaseComponentProps {
  variant?: ComponentVariant;
  size?: ComponentSize;
  type?: "button" | "submit" | "reset";
  href?: string;
  ripple?: boolean;
  fullWidth?: boolean;
  onclick?: (event: Event) => void;
}

export interface LoadingSpinnerProps extends BaseComponentProps {
  size?: ComponentSize;
  variant?: ComponentVariant;
  message?: string;
  overlay?: boolean;
  overlayColor?: string;
}
