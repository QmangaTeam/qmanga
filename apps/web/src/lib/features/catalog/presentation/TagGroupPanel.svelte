<script lang="ts">
	import type { Genre } from '../application/catalog';
	import { useLocale, text } from '$lib/shared/presentation/locale';
	import Icon from '$lib/shared/presentation/ui/Icon.svelte';

	let { name, items, selected = [], limit = 12, ontoggle, onback }: { name: string; items: Genre[]; selected?: string[]; limit?: number; ontoggle: (slug: string) => void; onback: () => void } = $props();
	const { locale } = useLocale();
	const instance = $props.id();
	const inputId = 'tag-group-search-' + instance;
	let query = $state('');
	function label(item: Genre): string {
		return $locale === 'ru' && item.nameRu ? item.nameRu : item.name;
	}
	const needle = $derived(query.trim().toLowerCase());
	const visible = $derived(needle ? items.filter(item => [item.name, item.nameRu || '', item.slug].join(' ').toLowerCase().includes(needle)) : items);
	const full = $derived(selected.length >= limit);
</script>

<section class="tag-group-panel" aria-label={name}>
	<div class="panel-head">
		<button class="back" type="button" aria-label={text($locale, 'Артқа', 'Назад')} onclick={onback}><Icon name="left" /></button>
		<h3>{name}</h3>
	</div>
	<div class="panel-search">
		<label class="search-label" for={inputId}>{text($locale, 'Топ ішінен іздеу', 'Поиск по группе')}</label>
		<div class="search-field">
			<span class="search-icon"><Icon name="search" size={18} /></span>
			<input id={inputId} type="search" bind:value={query} placeholder={text($locale, 'Тег атауы', 'Название тега')} autocomplete="off" spellcheck="false" />
			{#if query}<button class="search-clear" type="button" aria-label={text($locale, 'Іздеуді тазарту', 'Очистить поиск')} onclick={() => (query = '')}><Icon name="close" size={16} /></button>{/if}
		</div>
	</div>
	{#if full}<p class="tag-notice" role="status">{text($locale, '12 тегтен көп таңдауға болмайды', 'Можно выбрать не более 12 тегов')}</p>{/if}
	<ul>
		{#each visible as item}
			<li>
				<label class="tag-option">
					<input type="checkbox" value={item.slug} checked={selected.includes(item.slug)} disabled={!selected.includes(item.slug) && full} onchange={() => ontoggle(item.slug)} />
					<span class:child={item.parentSlug}>{#if item.parentSlug}<span class="child-mark">—</span>{/if}{label(item)}</span>
				</label>
			</li>
		{/each}
	</ul>
	{#if !visible.length}<p class="tag-empty">{text($locale, 'Тег табылмады', 'Теги не найдены')}</p>{/if}
</section>

<style>
	.tag-group-panel { display: flex; flex-direction: column; min-width: 0; min-height: 0; background: var(--panel); }
	.panel-head { display: flex; align-items: center; gap: 10px; padding: 8px 12px 8px 8px; border-bottom: 1px solid var(--line); }
	.panel-head h3 { margin: 0; font-size: 16px; line-height: 1.3; font-weight: 700; overflow-wrap: anywhere; }
	.back { display: flex; align-items: center; justify-content: center; width: 44px; height: 44px; flex: none; border: 0; border-radius: 999px; background: var(--raised); color: var(--text); cursor: pointer; }
	.back:hover { color: var(--accent); }
	.back:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
	.panel-search { display: grid; gap: 6px; padding: 12px 16px 8px; }
	.search-label { font-size: 12px; color: var(--muted); }
	.search-field { position: relative; display: flex; align-items: center; min-width: 0; }
	.search-field input { width: 100%; min-width: 0; min-height: 44px; padding: 10px 44px 10px 40px; border: 1px solid var(--line); border-radius: 999px; background: var(--bg); color: var(--text); font: inherit; font-size: 15px; line-height: 1.4; }
	.search-field input:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
	.search-field input::-webkit-search-cancel-button { -webkit-appearance: none; appearance: none; display: none; }
	.search-icon { position: absolute; left: 14px; display: flex; color: var(--muted); pointer-events: none; }
	.search-clear { position: absolute; right: 2px; display: flex; align-items: center; justify-content: center; width: 40px; height: 40px; border: 0; border-radius: 999px; background: none; color: var(--muted); cursor: pointer; }
	.search-clear:hover { color: var(--text); }
	.tag-notice, .tag-empty { margin: 0; padding: 0 16px 6px; font-size: 12px; line-height: 1.4; color: var(--accent); }
	.tag-empty { color: var(--muted); font-size: 13px; }
	ul { flex: 1; min-height: 0; margin: 0; padding: 0 16px 16px; list-style: none; overflow-y: auto; overscroll-behavior: contain; }
	.tag-option { display: flex; align-items: center; gap: 10px; min-height: 44px; font-size: 14px; line-height: 1.35; color: var(--text); cursor: pointer; overflow-wrap: anywhere; }
	.tag-option input { width: 20px; height: 20px; flex: none; margin: 0; accent-color: var(--accent); }
	.tag-option input:disabled { opacity: .45; }
	.tag-option span.child { color: var(--muted); }
	.child-mark { margin-right: 4px; }
</style>
