<script lang="ts">
 import type { ActionLinkProps } from './types';
 let { href, variant = 'secondary', class: className = '', children, ...attributes }: ActionLinkProps = $props();
</script>

<a {...attributes} {href} class={`ui-action ui-action--${variant} ${className}`}>{@render children?.()}</a>
