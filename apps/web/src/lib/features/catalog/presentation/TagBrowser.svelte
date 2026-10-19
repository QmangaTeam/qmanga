<script lang="ts">
	import type { Genre } from '../application/catalog';
	import { useLocale, text } from '$lib/shared/presentation/locale';
	import Icon from '$lib/shared/presentation/ui/Icon.svelte';

	let { genres, categories = [], selected = [], limit = 12, ontoggle, onopen }: { genres: Genre[]; categories?: Genre[]; selected?: string[]; limit?: number; ontoggle: (slug: string) => void; onopen: (slug: string) => void } = $props();
	const { locale } = useLocale();
	const instance = $props.id();
	const inputId = 'tag-search-' + instance;
	let query = $state('');

	function name(item: Genre): string {
		return $locale === 'ru' && item.nameRu ? item.nameRu : item.name;
	}
	function groupName(slug: string): string {
		const group = categories.find(item => item.slug === slug);
		if (!group) return text($locale, 'Басқа', 'Прочее');
		return $locale === 'ru' && group.nameRu ? group.nameRu : group.name;
	}
	function matches(item: Genre, needle: string): boolean {
		if (!needle) return true;
		return [item.name, item.nameRu || '', item.slug].join(' ').toLowerCase().includes(needle);
	}
	const needle = $derived(query.trim().toLowerCase());
	const groups = $derived.by(() => {
		const buckets = new Map<string, Genre[]>();
		for (const item of genres) {
			if (!matches(item, needle)) continue;
			const key = item.categorySlug || '';
			const bucket = buckets.get(key) || [];
			bucket.push(item);
			buckets.set(key, bucket);
		}
		const order = categories.map(group => group.slug).filter(slug => buckets.has(slug));
		if (buckets.has('')) order.push('');
		return order.map(slug => ({ slug, name: groupName(slug), items: buckets.get(slug) || [], popular: !needle && (buckets.get(slug) || []).length > 5 }));
	});
	const full = $derived(selected.length >= limit);
</script>

<div class="tag-browser">
	<label class="search-label" for={inputId}>{text($locale, 'Тег іздеу', 'Поиск тегов')}</label>
	<div class="search-field">
		<span class="search-icon"><Icon name="search" size={18} /></span>
		<input id={inputId} type="search" bind:value={query} placeholder={text($locale, 'Мысалы: романтика', 'Например: романтика')} autocomplete="off" spellcheck="false" />
		{#if query}<button class="search-clear" type="button" aria-label={text($locale, 'Іздеуді тазарту', 'Очистить поиск')} onclick={() => (query = '')}><Icon name="close" size={16} /></button>{/if}
	</div>
	<p class="tag-hint">{text($locale, 'Барлық таңдалған тег сәйкес келуі керек', 'Тайтл должен иметь все выбранные теги')}</p>
	{#if full}<p class="tag-notice" role="status">{text($locale, '12 тегтен көп таңдауға болмайды', 'Можно выбрать не более 12 тегов')}</p>{/if}
	{#each groups as group}
		<section class="tag-group" data-group={group.slug}>
			<div class="tag-group-head">
				<h3>{group.name}</h3>
				{#if !needle}<button class="group-open" type="button" aria-label={text($locale, 'Топтың барлық тегтері', 'Все теги группы') + ': ' + group.name} onclick={() => onopen(group.slug)}><Icon name="right" size={18} /></button>{/if}
			</div>
			{#if group.popular}<p class="tag-subtitle">{text($locale, 'Танымал', 'Популярные')}</p>{/if}
			<ul>
				{#each (group.popular ? group.items.slice(0, 5) : group.items) as item}
					<li>
						<label class="tag-option">
							<input type="checkbox" value={item.slug} checked={selected.includes(item.slug)} disabled={!selected.includes(item.slug) && full} onchange={() => ontoggle(item.slug)} />
							<span class:child={item.parentSlug}>{#if item.parentSlug}<span class="child-mark">—</span>{/if}{name(item)}</span>
						</label>
					</li>
				{/each}
			</ul>
			{#if group.popular}<button class="group-more" type="button" onclick={() => onopen(group.slug)}>{text($locale, 'Барлығы', 'Полностью')}<Icon name="right" size={16} /></button>{/if}
		</section>
	{/each}
	{#if !groups.length}<p class="tag-empty">{text($locale, 'Тег табылмады', 'Теги не найдены')}</p>{/if}
</div>

<style>
	.tag-browser { display: grid; gap: 12px; min-width: 0; }
	.search-label { font-size: 13px; color: var(--muted); }
	.search-field { position: relative; display: flex; align-items: center; min-width: 0; }
	.search-field input { width: 100%; min-width: 0; min-height: 44px; padding: 10px 44px 10px 40px; border: 1px solid var(--line); border-radius: 999px; background: var(--panel); color: var(--text); font: inherit; font-size: 15px; line-height: 1.4; }
	.search-field input:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
	.search-field input::-webkit-search-cancel-button { -webkit-appearance: none; appearance: none; display: none; }
	.search-icon { position: absolute; left: 14px; display: flex; color: var(--muted); pointer-events: none; }
	.search-clear { position: absolute; right: 2px; display: flex; align-items: center; justify-content: center; width: 40px; height: 40px; border: 0; border-radius: 999px; background: none; color: var(--muted); cursor: pointer; }
	.search-clear:hover { color: var(--text); }
	.search-clear:focus-visible { outline: 2px solid var(--accent); outline-offset: 1px; }
	.tag-hint, .tag-notice, .tag-empty { margin: 0; font-size: 12px; line-height: 1.4; color: var(--muted); }
	.tag-notice { color: var(--accent); }
	.tag-empty { padding: 8px 0; font-size: 13px; }
	.tag-group { display: grid; gap: 6px; min-width: 0; }
	.tag-group + .tag-group { padding-top: 12px; border-top: 1px solid var(--line); }
	.tag-group-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; min-height: 44px; }
	.tag-group-head h3 { margin: 0; font-size: 14px; line-height: 1.3; font-weight: 700; }
	.group-open { display: flex; align-items: center; justify-content: center; width: 44px; height: 44px; flex: none; border: 0; border-radius: 999px; background: var(--raised); color: var(--text); cursor: pointer; }
	.group-open:hover { color: var(--accent); }
	.group-open:focus-visible, .group-more:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
	.tag-subtitle { margin: 0; font-size: 12px; font-weight: 600; color: var(--muted); }
	ul { display: grid; gap: 1px; margin: 0; padding: 0; list-style: none; }
	.tag-option { display: flex; align-items: center; gap: 10px; min-height: 44px; font-size: 14px; line-height: 1.35; color: var(--text); cursor: pointer; overflow-wrap: anywhere; }
	.tag-option input { width: 20px; height: 20px; flex: none; margin: 0; accent-color: var(--accent); }
	.tag-option input:disabled { opacity: .45; }
	.tag-option span.child { color: var(--muted); }
	.child-mark { margin-right: 4px; }
	.group-more { display: inline-flex; align-items: center; justify-content: center; gap: 6px; min-height: 44px; padding: 4px 8px; border: 0; border-radius: 8px; background: none; color: var(--muted); font: inherit; font-size: 13px; cursor: pointer; }
	.group-more:hover { color: var(--accent); }
</style>
