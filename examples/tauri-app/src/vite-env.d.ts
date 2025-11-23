/// <reference types="svelte" />
/// <reference types="vite/client" />

// Svelte 5 runes type declarations for JavaScript files
declare function $state<T>(initial: T): T;
declare function $state<T>(): T | undefined;
declare function $derived<T>(expression: T): T;
declare function $effect(fn: () => void | (() => void)): void;
declare function $props<T>(): T;

// Svelte HTML namespace for element types
declare namespace svelteHTML {
  interface HTMLAttributes<T> {
    [key: string]: any;
  }
}
