<script lang="ts">
 let { label, tone = 'neutral' }: { label: string; tone?: 'neutral'|'accent'|'warning'|'danger' } = $props();
</script>
<span class="badge" class:accent={tone === 'accent'} class:emphasis={tone === 'warning' || tone === 'danger'}>{label}</span>
<style>
 .badge { display:inline-flex; align-items:center; padding:2px 8px; border-radius:4px; background:var(--raised); color:var(--muted); font-size:13px; line-height:1.5; overflow-wrap:anywhere; }
 .accent { color:var(--accent); } .emphasis { color:var(--text); border:1px solid var(--muted); font-weight:600; }
</style>
