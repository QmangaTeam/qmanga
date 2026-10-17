<script lang="ts">
 import type { ButtonProps } from './types';
 let { type = 'button', variant = 'secondary', busy = false, disabled = false, class: className = '', children, ...attributes }: ButtonProps = $props();
</script>

<button {...attributes} {type} disabled={disabled || busy} aria-busy={busy ? 'true' : undefined} class={`ui-action ui-action--${variant} ${className}`}>
 {@render children?.()}
</button>
