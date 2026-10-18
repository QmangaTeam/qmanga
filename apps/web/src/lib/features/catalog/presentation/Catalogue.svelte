<script lang="ts">
 import type {CatalogFilter,Genre,ResultPage,Summary} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {catalogNames} from './messages';
 import CoverCard from './CoverCard.svelte';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import Button from '$lib/shared/presentation/ui/Button.svelte';
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 import Pager from '$lib/shared/presentation/ui/Pager.svelte';
 let {view}:{view:{available:boolean;message:string;filter:CatalogFilter;genres:Genre[];page:ResultPage<Summary>;nextURL:string|null}}=$props();
 const {locale}=useLocale();
 const names=$derived(catalogNames($locale));
 const active=$derived([
  view.filter.query&&text($locale,'Іздеу: ','Поиск: ')+view.filter.query,
  view.filter.kind&&names.kind[view.filter.kind as keyof typeof names.kind],
  view.filter.genre&&(view.genres.find(genre=>genre.slug===view.filter.genre)?.name||view.filter.genre),
  view.filter.originalStatus&&names.original[view.filter.originalStatus as keyof typeof names.original],
  view.filter.translationStatus&&names.translation[view.filter.translationStatus as keyof typeof names.translation],
  view.filter.hasChapters&&text($locale,'Тараулары бар','Есть главы'),
  view.filter.featured&&text($locale,'Ұсынылған','Рекомендуемые'),
  view.filter.sort==='name'&&text($locale,'Атауы бойынша','По названию'),
  view.filter.sort==='new'&&text($locale,'Жаңа тайтлдар','Новые тайтлы')
 ].filter(Boolean));
</script>
<svelte:head><title>Каталог — Qmanga</title><meta name="description" content={text($locale,'Қазақ тіліндегі манга, манхва және маньхуа.','Манга, манхва и маньхуа на казахском языке.')}/></svelte:head>
<SectionHeading level={1} title="Каталог"/>
<form action="/catalog" method="GET" class="catalogue-search" role="search">
 {#if view.filter.featured}<input type="hidden" name="featured" value="true"/>{/if}
 <label class="sr-only" for="catalogue-query">{text($locale,'Тайтл немесе басқа атауы','Тайтл или альтернативное название')}</label>
 <div class="search-row"><input id="catalogue-query" type="search" name="q" value={view.filter.query} maxlength="80" placeholder={text($locale,'Атауын іздеу','Поиск по названию')}/><Button type="submit" variant="primary">{text($locale,'Іздеу','Найти')}</Button></div>
 <details open={!!(view.filter.kind||view.filter.genre||view.filter.originalStatus||view.filter.translationStatus)}><summary>{text($locale,'Сүзгілер мен реттеу','Фильтры и сортировка')}</summary><div class="filter-fields">
 <label>{text($locale,'Түрі','Тип')}<select name="kind" value={view.filter.kind}><option value="">{text($locale,'Барлығы','Все')}</option>{#each Object.entries(names.kind) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>Жанр<select name="genre" value={view.filter.genre}><option value="">{text($locale,'Барлығы','Все')}</option>{#each view.genres as genre}<option value={genre.slug}>{genre.name}</option>{/each}</select></label>
 <label>{text($locale,'Түпнұсқа','Статус оригинала')}<select name="original_status" value={view.filter.originalStatus}><option value="">{text($locale,'Барлығы','Все')}</option>{#each Object.entries(names.original) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>{text($locale,'Аударма','Статус перевода')}<select name="translation_status" value={view.filter.translationStatus}><option value="">{text($locale,'Барлығы','Все')}</option>{#each Object.entries(names.translation) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>{text($locale,'Реттеу','Сортировка')}<select name="sort" value={view.filter.sort}><option value="updated">{text($locale,'Соңғы жаңартулар','Недавно обновлённые')}</option><option value="new">{text($locale,'Жаңа тайтлдар','Новые тайтлы')}</option><option value="name">{text($locale,'Атауы бойынша','По названию')}</option></select></label>
 <label class="published-check"><input type="checkbox" name="has_chapters" value="true" checked={view.filter.hasChapters}/>{text($locale,'Оқуға болатын тараулары бар','Есть опубликованные главы')}</label><div class="filter-actions"><Button type="submit" variant="primary">{text($locale,'Қолдану','Применить')}</Button><ActionLink href="/catalog">{text($locale,'Тазарту','Сбросить')}</ActionLink></div>
 </div></details>
</form>
{#if active.length}<div class="active-filters" aria-label={text($locale,'Қолданылған сүзгілер','Применённые фильтры')}>{#each active as label}<span>{label}</span>{/each}<ActionLink href="/catalog">{text($locale,'Барлығын тазарту','Сбросить всё')}</ActionLink></div>{/if}
{#if !view.available}<StateMessage kind="error" title={text($locale,'Каталог жүктелмеді','Каталог не загрузился')}>{#snippet actions()}<ActionLink href="/catalog">{text($locale,'Қайталау','Повторить')}</ActionLink>{/snippet}</StateMessage>
{:else if !view.page.items.length}<StateMessage kind="empty" title={text($locale,'Тайтл табылмады','Тайтлы не найдены')} description={text($locale,'Басқа атаумен ізде немесе сүзгілерді тазарт.','Попробуй другое название или сбрось фильтры.')}/>
{:else}<div class="cover-grid">{#each view.page.items as title,index}<CoverCard href={'/titles/'+encodeURIComponent(title.slug)} name={title.name} coverURL={title.coverURL} metadata={names.kind[title.kind]} tag={title.demo?text($locale,'Үлгі','Пример'):''} score={title.score} ratingCount={title.ratingCount} fallback={text($locale,'Мұқаба жоқ','Нет обложки')} priority={index<2}/>{/each}</div>{/if}
{#if view.nextURL}<Pager label={text($locale,'Каталог беттері','Страницы каталога')} next={{label:text($locale,'Келесі бет','Следующая страница'),href:view.nextURL}}/>{/if}
<style>.active-filters{display:flex;align-items:center;gap:8px;flex-wrap:wrap;margin:-12px 0 24px}.active-filters>span{font-size:13px;padding:6px 10px;background:var(--raised);border-radius:8px}.filter-fields .published-check{display:flex;align-items:center;gap:10px;min-height:44px}.published-check input{width:20px;height:20px;min-height:20px}.catalogue-search { margin-bottom:24px; } .search-row { display:flex; gap:8px; } .search-row input { flex:1; } summary { cursor:pointer; min-height:44px; padding:12px 0; font-size:14px; color:var(--muted); } .filter-fields { display:grid; gap:16px; margin:12px 0 24px; } .filter-fields label { display:grid; gap:8px; font-size:14px; } .filter-actions { display:flex; flex-wrap:wrap; align-items:end; gap:8px; } @media(min-width:640px) { .filter-fields { grid-template-columns:repeat(2,minmax(0,1fr)); } } @media(min-width:1024px) { .filter-fields { grid-template-columns:repeat(3,minmax(0,1fr)); } }</style>
