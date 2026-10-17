<script lang="ts">
 import Button from './Button.svelte';
 import Icon from './Icon.svelte';
 import type { ButtonProps, IconName } from './types';
 let { name, label, class: className = '', variant = 'ghost', ...attributes }: Omit<ButtonProps, 'children' | 'aria-label'> & { name: IconName; label: string } = $props();
</script>

<Button {...attributes} {variant} aria-label={label} title={label} class={`ui-icon-button ${className}`}><Icon {name}/></Button>
