<script lang="ts">
 import {tick} from 'svelte';
 import type {Genre} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {label,items,selected,onchange,limit}:{label:string;items:Genre[];selected:string[];onchange:(ids:string[])=>void;limit?:number}=$props();
 const {locale}=useLocale();
 const instance=$props.id();
 let open=$state(false),query=$state('');
 let trigger:HTMLButtonElement;
 let search=$state<HTMLInputElement>();
 const matches=$derived(items.filter(item=>[item.name,item.nameRu||'',item.slug].join(' ').toLowerCase().includes(query.trim().toLowerCase())));
 function name(item:Genre){return $locale==='ru'&&item.nameRu?item.nameRu:item.name;}
 function toggle(id:string,checked:boolean){if(checked&&(selected.includes(id)||(limit!==undefined&&selected.length>=limit)))return;onchange(checked?[...selected,id]:selected.filter(value=>value!==id));}
 async function show(){open=true;await tick();search?.focus();}
 async function close(){open=false;query='';await tick();trigger?.focus();}
</script>

<fieldset class="taxonomy-selection">
 <legend>{label}</legend>
 <div class="selection-heading"><span>{text($locale,'Таңдалған: ','Выбрано: ')}{selected.length}{limit!==undefined?' / '+limit:''}</span><button bind:this={trigger} type="button" class="button subtle" aria-expanded={open} aria-controls={'taxonomy-'+instance} onclick={()=>open?close():show()}>{open?text($locale,'Дайын','Готово'):text($locale,'Таңдау','Выбрать')}</button></div>
 {#if selected.length}<ul class="selected" aria-label={text($locale,'Таңдалған мәндер','Выбранные значения')}>{#each selected as id}{@const item=items.find(item=>item.id===id)}{@const itemName=item?name(item):id}<li><button type="button" aria-label={text($locale,'Алып тастау: ','Убрать: ')+itemName} onclick={()=>toggle(id,false)}>{itemName} <span aria-hidden="true">×</span></button></li>{/each}</ul>{/if}
 <div id={'taxonomy-'+instance} hidden={!open}>
 {#if open}
  <label class="search-label" for={'taxonomy-search-'+instance}>{text($locale,'Іздеу','Поиск')}<input bind:this={search} id={'taxonomy-search-'+instance} type="search" bind:value={query} maxlength="80" autocomplete="off" onkeydown={event=>{if(event.key==='Enter')event.preventDefault();if(event.key==='Escape'){event.preventDefault();event.stopPropagation();void close();}}}/></label>
  <p class="result-count" role="status">{text($locale,'Табылды: ','Найдено: ')}{matches.length}{#if limit!==undefined&&selected.length>=limit} · {text($locale,'Шекке жетті. Жаңасын қосу үшін біреуін алып таста.','Достигнут лимит. Убери выбранное, чтобы добавить другое.')}{/if}</p>
  <div class="options">
   {#each matches as item (item.id)}<label class="option"><input type="checkbox" value={item.id} checked={selected.includes(item.id)} disabled={!selected.includes(item.id)&&limit!==undefined&&selected.length>=limit} onchange={event=>toggle(item.id,event.currentTarget.checked)}/><span>{name(item)}</span></label>{:else}<p role="status">{text($locale,'Ештеңе табылмады','Ничего не найдено')}</p>{/each}
  </div>
 {/if}
 </div>
</fieldset>

<style>
 .taxonomy-selection{min-width:0;margin:0;padding:12px;border:1px solid var(--line);border-radius:8px}
 legend{font-size:14px;font-weight:600;padding-inline:4px}.selection-heading{display:flex;align-items:center;justify-content:space-between;gap:12px;font-size:13px;color:var(--muted)}
 .selected{display:flex;flex-wrap:wrap;gap:6px;max-height:144px;overflow:auto;list-style:none;margin:8px 0 0;padding:4px}
 .selected li{max-width:100%}.selected button{min-height:44px;display:flex;align-items:center;gap:8px;max-width:100%;border:1px solid var(--line);border-radius:8px;background:var(--raised);color:var(--text);font:inherit;font-size:13px;padding:8px;overflow-wrap:anywhere;text-align:left}
 .search-label{display:grid;gap:8px;margin-top:12px;font-size:14px}.options{display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,210px),1fr));gap:0 12px;max-height:260px;overflow:auto;margin-top:8px;padding:4px;overscroll-behavior:contain}
 .option{display:flex;align-items:center;gap:10px;min-height:44px;font-size:14px;overflow-wrap:anywhere;cursor:pointer}.option input{width:20px;height:20px;min-height:20px;margin:0;flex:none;accent-color:var(--accent)}
 .options p{font-size:14px;color:var(--muted)}
.result-count{font-size:12px;color:var(--muted);margin:12px 0 0}.option:has(input:disabled){opacity:.55}
</style>
