<script lang="ts">
 import {onMount,untrack} from 'svelte';
 import type {CatalogFilter,Genre,Collection} from '../application/catalog';
 import {ageRatings} from '../domain/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {catalogNames} from './messages';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {filter,genres,categories=[],collections=[]}:{filter:CatalogFilter;genres:Genre[];categories?:Genre[];collections?:Collection[]}=$props();
 const {locale}=useLocale();
 const names=$derived(catalogNames($locale));
 const heading=$derived(text($locale,'Каталог сүзгілері','Фильтры каталога'));
 function values(value:CatalogFilter){return {category:value.category||'',collection:value.collection||'',kind:value.kind||'',genre:value.genre||'',originalStatus:value.originalStatus||'',translationStatus:value.translationStatus||'',ageRating:value.ageRating||'',sort:value.sort||'updated',hasChapters:value.hasChapters===true};}
 let draft=$state(untrack(()=>values(filter)));
 let dialog:HTMLDialogElement|undefined;
 $effect(()=>{draft=values(filter);});
 function open(){draft=values(filter);dialog?.showModal();}
 function close(){dialog?.close();}
 function backdrop(event:MouseEvent){
  if(event.target!==dialog||!dialog)return;
  const bounds=dialog.getBoundingClientRect();
  if(event.clientX<bounds.left||event.clientX>bounds.right||event.clientY<bounds.top||event.clientY>bounds.bottom)close();
 }
 onMount(()=>{
  const desktop=window.matchMedia('(min-width:1024px)');
  const changed=()=>{if(desktop.matches)close();};
  desktop.addEventListener('change',changed);
  return ()=>{desktop.removeEventListener('change',changed);close();};
 });
</script>
{#snippet hiddenFields()}
 <input type="hidden" name="q" value={filter.query||''}/>
 {#if filter.featured}<input type="hidden" name="featured" value="true"/>{/if}
{/snippet}
{#snippet fields()}
 <div class="filter-fields">
  <label>{text($locale,'Түрі','Тип')}<select name="kind" bind:value={draft.kind}><option value="">{text($locale,'Барлығы','Все')}</option>{#each Object.entries(names.kind) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
  <label>{text($locale,'Мазмұн санаты','Категория содержания')}<select name="category" bind:value={draft.category}><option value="">{text($locale,'Барлығы','Все')}</option>{#each categories as item}<option value={item.slug}>{item.name}</option>{/each}</select></label>
  <label>{text($locale,'Редакциялық жинақ','Редакционная подборка')}<select name="collection" bind:value={draft.collection}><option value="">{text($locale,'Барлығы','Все')}</option>{#each collections as item}<option value={item.slug}>{item.name}</option>{/each}</select></label>
  <label>Жанр<select name="genre" bind:value={draft.genre}><option value="">{text($locale,'Барлығы','Все')}</option>{#each genres as genre}<option value={genre.slug}>{genre.name}</option>{/each}</select></label>
  <label>{text($locale,'Түпнұсқа','Статус оригинала')}<select name="original_status" bind:value={draft.originalStatus}><option value="">{text($locale,'Барлығы','Все')}</option>{#each Object.entries(names.original) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
  <label>{text($locale,'Аударма','Статус перевода')}<select name="translation_status" bind:value={draft.translationStatus}><option value="">{text($locale,'Барлығы','Все')}</option>{#each Object.entries(names.translation) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
  <label>{text($locale,'Жас санаты','Возрастная маркировка')}<select name="age_rating" bind:value={draft.ageRating}><option value="">{text($locale,'Барлығы','Все')}</option>{#each ageRatings as age}<option value={String(age)}>{age===0?text($locale,'Көрсетілмеген','Не указано'):age+'+'}</option>{/each}</select></label>
  <label>{text($locale,'Реттеу','Сортировка')}<select name="sort" bind:value={draft.sort}><option value="updated">{text($locale,'Соңғы жаңартулар','Недавно обновлённые')}</option><option value="new">{text($locale,'Жаңа тайтлдар','Новые тайтлы')}</option><option value="name">{text($locale,'Атауы бойынша','По названию')}</option></select></label>
  <label class="published-check"><input type="checkbox" name="has_chapters" value="true" bind:checked={draft.hasChapters}/>{text($locale,'Оқуға болатын тараулары бар','Есть опубликованные главы')}</label>
 </div>
{/snippet}
{#snippet actions()}
 <div class="filter-actions"><button class="button primary apply-filters" type="submit">{text($locale,'Қолдану','Применить')}</button><a class="filter-reset" href="/catalog" onclick={close}>{text($locale,'Тазарту','Сбросить')}</a></div>
{/snippet}
<div class="filters-shell">
 <button class="button subtle filter-trigger" type="button" onclick={open} aria-haspopup="dialog"><Icon name="settings" size={18}/>{text($locale,'Сүзгілер','Фильтры')}</button>
 <aside class="catalogue-filters" aria-label={heading}>
  <h2>{heading}</h2><form action="/catalog" method="GET">{@render hiddenFields()}{@render fields()}{@render actions()}</form>
 </aside>
 <dialog bind:this={dialog} class="catalogue-filter-dialog" aria-labelledby="catalogue-filter-title" onclick={backdrop}>
  <form action="/catalog" method="GET" onsubmit={close}>
   <div class="filter-heading"><h2 id="catalogue-filter-title">{heading}</h2><button class="close-filters" type="button" aria-label={text($locale,'Сүзгілерді жабу','Закрыть фильтры')} onclick={close}><Icon name="close"/></button></div>
   {@render hiddenFields()}{@render fields()}{@render actions()}
  </form>
 </dialog>
</div>
<style>
 .filters-shell{min-width:0;grid-column:2;grid-row:1}
 .filter-trigger{display:inline-flex;align-items:center;gap:8px;margin:0;min-height:44px}
 .catalogue-filters{display:none}
 h2{font-size:17px;line-height:1.35;margin:0;font-weight:700}
 .filter-fields{display:grid;gap:14px;min-height:0;min-width:0;padding:0 20px 16px;overflow-y:auto;overscroll-behavior:contain}
 .filter-fields label{display:grid;gap:7px;min-width:0;font-size:13px;color:var(--muted)}
 .filter-fields select{color:var(--text);min-height:44px}
 .filter-fields .published-check{display:flex;align-items:center;gap:10px;min-height:44px}
 .published-check input{width:20px;height:20px;min-height:20px;flex:none}
 .filter-actions{display:grid;gap:4px;flex-shrink:0;padding:16px 20px max(16px,env(safe-area-inset-bottom));background:var(--panel);border-top:1px solid var(--line)}
 .apply-filters{width:100%;min-height:44px}
 .filter-reset{display:flex;align-items:center;justify-content:center;min-height:44px;color:var(--muted);font-size:14px;text-decoration:none}
 .catalogue-filter-dialog{position:fixed;inset:0 0 0 auto;margin:0;width:min(360px,100vw);height:100dvh;max-height:100dvh;max-width:100vw;padding:0;border:0;border-left:1px solid var(--line);border-radius:14px 0 0 14px;background:var(--panel);color:var(--text);box-shadow:-16px 0 64px #0006;overflow:auto;overscroll-behavior:contain}
 .catalogue-filter-dialog[open]{animation:filter-enter 220ms cubic-bezier(.22,1,.36,1)}
 .catalogue-filter-dialog::backdrop{background:#0009;animation:filter-dim 180ms ease}
 .catalogue-filter-dialog form{display:flex;flex-direction:column;height:100%;min-height:0}
 .catalogue-filter-dialog .filter-fields{flex:1}
 .filter-heading{display:flex;align-items:center;justify-content:space-between;gap:12px;flex-shrink:0;padding:12px 12px 16px 20px}
 .close-filters{display:flex;align-items:center;justify-content:center;min-width:44px;min-height:44px;border:0;border-radius:8px;background:none;color:var(--text)}
 @keyframes filter-enter{from{transform:translateX(100%)}to{transform:translateX(0)}}
 @keyframes filter-dim{from{opacity:0}to{opacity:1}}
 @media(prefers-reduced-motion:reduce){.catalogue-filter-dialog[open],.catalogue-filter-dialog::backdrop{animation:none}}
 @media(min-width:1024px){
  .filters-shell{grid-column:2;grid-row:1/span 2;align-self:start;position:sticky;top:88px}
  .filter-trigger{display:none}
  .catalogue-filters{display:block;border:1px solid var(--line);border-radius:12px;background:var(--panel);overflow:hidden}
  .catalogue-filters h2{padding:20px 20px 16px}
  .catalogue-filters form{display:flex;flex-direction:column;max-height:calc(100dvh - 166px);min-height:180px}
  .catalogue-filters .filter-fields{flex:1}
 }
</style>
