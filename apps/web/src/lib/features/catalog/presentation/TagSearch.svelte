<script lang="ts">
 import type {Genre} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';

 type TagSearchProps={genres:Genre[];categories?:Genre[];selected?:string[];onchange:(slugs:string[])=>void;field?:string};
 let {genres,categories=[],selected=[],onchange,field='genre'}:TagSearchProps=$props();
 const {locale}=useLocale();
 const instance=Math.random().toString(36).slice(2,9);
 const inputId='tag-search-'+instance;
 const listId='tag-list-'+instance;
 let query=$state('');
 let open=$state(false);

 function name(item:Genre):string {
  const russian=item.nameRu||'';
  return $locale==='ru'&&russian?russian:item.name;
 }
 function groupName(slug:string):string {
  const group=categories.find(item=>item.slug===slug);
  if(!group)return slug;
  const russian=group.nameRu||'';
  return $locale==='ru'&&russian?russian:group.name;
 }
 function matches(item:Genre,needle:string):boolean {
  if(!needle)return true;
  const haystack=[item.name,item.nameRu||'',item.slug].join(' ').toLowerCase();
  return haystack.includes(needle);
 }
 const needle=$derived(query.trim().toLowerCase());
 const suggestions=$derived.by(()=>{
  const limit=needle?40:6;
  const groups=new Map<string,Genre[]>();
  for(const item of genres){
   if(!matches(item,needle))continue;
   const key=item.categorySlug||'';
   const bucket=groups.get(key)||[];
   if(bucket.length<limit)bucket.push(item);
   groups.set(key,bucket);
  }
  return [...groups.entries()].map(([slug,items])=>({slug,items}));
 });
 function choose(item:Genre){
  if(!selected.includes(item.slug))onchange([...selected,item.slug]);
  query='';
 }
 function remove(slug:string){
  onchange(selected.filter(item=>item!==slug));
 }
 function onkeydown(event:KeyboardEvent){
  if(event.key==='Escape'){open=false;return;}
  if(event.key==='Enter'){
   const first=suggestions[0]?.items[0];
   if(first&&open){event.preventDefault();choose(first);}
  }
 }
</script>
<div class="tag-search">
 <label class="tag-search-label" for={inputId}>{text($locale,'Тегтер','Теги')}</label>
 <div class="tag-search-field">
  <input id={inputId} type="search" bind:value={query} placeholder={text($locale,'Тег іздеу','Поиск тега')} autocomplete="off" role="combobox" aria-expanded={open} aria-controls={listId} aria-autocomplete="list" onfocus={()=>open=true} oninput={()=>open=true} onkeydown={onkeydown}/>
  {#if query}<button class="tag-search-clear" type="button" aria-label={text($locale,'Іздеуді тазарту','Очистить поиск')} onclick={()=>{query='';}}><Icon name="close" size={16}/></button>{/if}
 </div>
 {#if selected.length}
 <ul class="tag-search-selected" aria-label={text($locale,'Таңдалған тегтер','Выбранные теги')}>
  {#each selected as slug}
   <li><button type="button" data-slug={slug} onclick={()=>remove(slug)}>{name(genres.find(item=>item.slug===slug)||{id:'',name:slug,slug})}<Icon name="close" size={14}/></button></li>
  {/each}
 </ul>
 {/if}
 {#if open}
 <div class="tag-search-results" id={listId} role="listbox" aria-label={text($locale,'Тегтер','Теги')}>
  {#each suggestions as group}
   <div class="tag-search-group">
    <p class="tag-search-group-name">{groupName(group.slug)}</p>
    <ul>
     {#each group.items as item}
      <li><button type="button" role="option" aria-selected={selected.includes(item.slug)} data-slug={item.slug} onclick={()=>choose(item)}>
       {#if item.parentSlug}<span class="tag-search-child">—</span>{/if}
       <span>{name(item)}</span>
       {#if selected.includes(item.slug)}<Icon name="check" size={14}/>{/if}
      </button></li>
     {/each}
    </ul>
   </div>
  {/each}
  {#if !suggestions.length}<p class="tag-search-empty">{text($locale,'Тег табылмады','Теги не найдены')}</p>{/if}
 </div>
 {/if}
 {#each field?selected:[] as slug}<input type="hidden" name={field} value={slug}/>{/each}
</div>
<style>
 .tag-search{display:grid;gap:7px;min-width:0}
 .tag-search-label{font-size:13px;color:var(--muted)}
 .tag-search-field{position:relative;display:flex;align-items:center;min-width:0}
 .tag-search-field input{width:100%;min-width:0;min-height:44px;padding:10px 44px 10px 12px;border:1px solid var(--muted);border-radius:8px;background:var(--panel);color:var(--text);font:inherit;font-size:16px;line-height:1.4}
 .tag-search-field input:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
 .tag-search-clear{position:absolute;right:2px;display:flex;align-items:center;justify-content:center;width:40px;height:40px;border:0;border-radius:8px;background:none;color:var(--muted);cursor:pointer}
 .tag-search-clear:hover{color:var(--text)}
 .tag-search-clear:focus-visible{outline:2px solid var(--accent);outline-offset:1px}
 .tag-search-selected{display:flex;flex-wrap:wrap;gap:6px;margin:0;padding:0;list-style:none}
 .tag-search-selected button{display:inline-flex;align-items:center;gap:6px;min-height:32px;max-width:100%;padding:5px 10px;border:1px solid var(--line);border-radius:999px;background:var(--raised);color:var(--text);font:inherit;font-size:13px;line-height:1.3;cursor:pointer;overflow-wrap:anywhere}
 .tag-search-selected button:hover{color:var(--accent)}
 .tag-search-selected button:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
 .tag-search-results{max-height:260px;overflow-y:auto;overscroll-behavior:contain;border:1px solid var(--line);border-radius:10px;background:var(--bg)}
 .tag-search-group{padding:6px 0}
 .tag-search-group+.tag-search-group{border-top:1px solid var(--line)}
 .tag-search-group-name{margin:0;padding:6px 12px 4px;font-size:12px;font-weight:600;color:var(--muted)}
 .tag-search-group ul{display:grid;gap:1px;margin:0;padding:0;list-style:none}
 .tag-search-group button{display:flex;align-items:center;gap:6px;width:100%;min-height:44px;padding:8px 12px;border:0;background:none;color:var(--text);font:inherit;font-size:14px;line-height:1.35;text-align:left;cursor:pointer;overflow-wrap:anywhere}
 .tag-search-group button:hover{background:var(--raised)}
 .tag-search-group button:focus-visible{outline:2px solid var(--accent);outline-offset:-2px}
 .tag-search-group button[aria-selected=true]{color:var(--accent)}
 .tag-search-child{color:var(--muted)}
 .tag-search-empty{margin:0;padding:12px;font-size:13px;color:var(--muted)}
</style>
