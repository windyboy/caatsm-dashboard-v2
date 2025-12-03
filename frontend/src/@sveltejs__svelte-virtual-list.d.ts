declare module "@sveltejs/svelte-virtual-list" {
  import type { SvelteComponent } from "svelte";

  interface VirtualListProps<T = any> {
    items: T[];
    height?: string;
    itemHeight?: number;
    start?: number;
    end?: number;
    class?: string;
  }

  export default class VirtualList<T = any> extends SvelteComponent<
    VirtualListProps<T>
  > {}
}

