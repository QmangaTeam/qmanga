<script lang="ts">
 import type { Snippet } from 'svelte';
 import type { FieldControl } from './types';
 let { id, label, hint = '', error = '', required = false, disabled = false, children }: {
  id: string; label: string; hint?: string; error?: string; required?: boolean; disabled?: boolean; children: Snippet<[FieldControl]>;
 } = $props();
 const describedBy = $derived([hint ? id + '-hint' : '', error ? id + '-error' : ''].filter(Boolean).join(' ') || undefined);
</script>

<div class="ui-field">
 <label for={id}>{label}{#if required}<span aria-hidden="true"> *</span>{/if}</label>
 {@render children({ id, required, disabled, 'aria-describedby': describedBy, 'aria-invalid': error ? 'true' : undefined })}
 {#if hint}<p id={id + '-hint'} class="hint">{hint}</p>{/if}
 {#if error}<p id={id + '-error'} class="error">{error}</p>{/if}
</div>

<style>
 .ui-field { display: grid; gap: 8px; min-width: 0; }
 label { font-size: 14px; font-weight: 600; line-height: 1.4; overflow-wrap: anywhere; }
 .hint, .error { margin: 0; font-size: 13px; line-height: 1.5; overflow-wrap: anywhere; }
 .hint { color: var(--muted); }
 .error { border-inline-start: 2px solid var(--text); padding-inline-start: 8px; }
 .ui-field :global(input), .ui-field :global(select), .ui-field :global(textarea) { box-sizing: border-box; min-width: 0; width: 100%; min-height: 44px; border: 1px solid var(--muted); border-radius: 8px; padding: 10px 12px; background: var(--panel); color: var(--text); font: inherit; font-size: 16px; }
 .ui-field :global(textarea) { min-height: 112px; resize: vertical; }
 .ui-field :global([aria-invalid="true"]) { border: 2px solid var(--text); padding: 9px 11px; }
 .ui-field :global(:disabled) { cursor: default; opacity: .6; }
 .ui-field :global(:focus-visible) { outline: 2px solid var(--accent); outline-offset: 3px; }
</style>
