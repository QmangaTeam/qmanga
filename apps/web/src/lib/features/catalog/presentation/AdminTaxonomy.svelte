<script lang="ts">
 import {onMount} from 'svelte';
 import type {CatalogManagementGateway} from '../application/management';
 import type {Genre,Collection,Taxonomy,Summary} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 let {gateway}:{gateway:CatalogManagementGateway}=$props();
 const {locale}=useLocale();
 let taxonomy=$state<Taxonomy>({genres:[],categories:[],collections:[]});
 let kind=$state<'genres'|'categories'>('genres'), taxon=$state<Genre>({id:'',name:'',slug:''});
 const emptyCollection=():Collection=>({id:'',name:'',slug:'',description:'',position:0,visible:false,titleIds:[]});
 let collection=$state<Collection>(emptyCollection()), titles=$state<Summary[]>([]),more=$state(false),busy=$state(false),failure=$state(''),notice=$state('');
 async function reload(){taxonomy=await gateway.taxonomy();}
 async function operation(action:()=>Promise<void>){if(busy)return;busy=true;failure='';notice='';try{await action();notice='saved';}catch(error){failure=failureCode(error);}finally{busy=false;}}
 async function loadTitles(){const next=await gateway.titles(titles.length);titles=[...titles,...next];more=next.length===100;}
 onMount(()=>{void operation(async()=>{await reload();await loadTitles();});});
 function chooseTaxon(id:string){taxon={...(taxonomy[kind].find(item=>item.id===id)||{id:'',name:'',slug:''})};}
 function chooseCollection(id:string){const item=taxonomy.collections.find(item=>item.id===id);collection=item?{...item,titleIds:[...item.titleIds]}:emptyCollection();}
 function toggleTitle(id:string,checked:boolean){collection.titleIds=checked?[...collection.titleIds,id]:collection.titleIds.filter(value=>value!==id);}
 function move(index:number,direction:number){const next=[...collection.titleIds],target=index+direction;if(target<0||target>=next.length)return;[next[index],next[target]]=[next[target],next[index]];collection.titleIds=next;}
</script>
<svelte:head><title>{text($locale,'Жанрлар мен жинақтар','Жанры и подборки')} — Qmanga</title></svelte:head>
<h1>{text($locale,'Жанрлар, санаттар және жинақтар','Жанры, категории и подборки')}</h1>
<p class="muted">{text($locale,'Жанр — сүзгі. Санат — мазмұн белгісі. Жинақ — редакция таңдаған тайтлдар. Жас санаты тайтлда бөлек көрсетіледі.','Жанр служит фильтром, категория отмечает содержание, подборка объединяет выбранные редакцией тайтлы. Возраст указывается отдельно в тайтле.')}</p>
{#if failure}<p role="alert">{noticeText($locale,failure)}</p>{/if}{#if notice}<p role="status">{noticeText($locale,notice)}</p>{/if}
<div class="workspace-grid"><section class="panel"><h2>{text($locale,'Жанрлар мен санаттар','Жанры и категории')}</h2>
<form onsubmit={event=>{event.preventDefault();void operation(async()=>{await gateway.saveTaxon(kind,taxon);taxon={id:'',name:'',slug:''};await reload();});}}>
<label>{text($locale,'Түрі','Тип')}<select bind:value={kind} onchange={()=>{taxon={id:'',name:'',slug:''};}}><option value="genres">{text($locale,'Жанр','Жанр')}</option><option value="categories">{text($locale,'Мазмұн санаты','Категория содержания')}</option></select></label>
<label>{text($locale,'Таңдау','Выбрать')}<select value={taxon.id} onchange={event=>chooseTaxon(event.currentTarget.value)}><option value="">{text($locale,'Жаңа','Создать')}</option>{#each taxonomy[kind] as item}<option value={item.id}>{item.name}</option>{/each}</select></label>
<label>{text($locale,'Атауы','Название')}<input bind:value={taxon.name} required maxlength="80"/></label><label>{text($locale,'Сілтеме атауы','Имя в ссылке')}<input bind:value={taxon.slug} required maxlength="64" pattern="[a-z0-9]+(-[a-z0-9]+)*"/></label>
<button class="button primary" disabled={busy}>{text($locale,'Сақтау','Сохранить')}</button>
{#if taxon.id}<button class="button subtle" type="button" disabled={busy} onclick={()=>operation(async()=>{await gateway.deleteTaxon(kind,taxon.id);taxon={id:'',name:'',slug:''};await reload();})}>{text($locale,'Жою','Удалить')}</button>{/if}
</form></section>
<section class="panel"><h2>{text($locale,'Редакциялық жинақтар','Редакционные подборки')}</h2>
<form onsubmit={event=>{event.preventDefault();void operation(async()=>{await gateway.saveCollection(collection);collection=emptyCollection();await reload();});}}>
<label>{text($locale,'Таңдау','Выбрать')}<select value={collection.id} onchange={event=>chooseCollection(event.currentTarget.value)}><option value="">{text($locale,'Жаңа жинақ','Создать подборку')}</option>{#each taxonomy.collections as item}<option value={item.id}>{item.name}</option>{/each}</select></label>
<label>{text($locale,'Атауы','Название')}<input bind:value={collection.name} required maxlength="80"/></label><label>{text($locale,'Сілтеме атауы','Имя в ссылке')}<input bind:value={collection.slug} required maxlength="64" pattern="[a-z0-9]+(-[a-z0-9]+)*"/></label>
<label>{text($locale,'Сипаттама','Описание')}<textarea bind:value={collection.description} maxlength="4000" rows="4"></textarea></label><label>{text($locale,'Реті','Порядок')}<input type="number" bind:value={collection.position} min="0" max="1000000" required/></label>
<label class="check"><input type="checkbox" bind:checked={collection.visible}/>{text($locale,'Каталогта көрінеді','Видна в каталоге')}</label>
<h3>{text($locale,'Тайтлдардың реті','Порядок тайтлов')}</h3>
<ol>{#each collection.titleIds as id,index}<li><span>{titles.find(item=>item.id===id)?.name||id}</span><button type="button" disabled={index===0} onclick={()=>move(index,-1)} aria-label={text($locale,'Жоғары жылжыту','Поднять')}>↑</button><button type="button" disabled={index===collection.titleIds.length-1} onclick={()=>move(index,1)} aria-label={text($locale,'Төмен жылжыту','Опустить')}>↓</button><button type="button" onclick={()=>toggleTitle(id,false)} aria-label={text($locale,'Жинақтан алып тастау','Убрать из подборки')}>×</button></li>{/each}</ol>
<fieldset><legend>{text($locale,'Тайтлдарды таңдау','Выбрать тайтлы')}</legend>{#each titles as title}<label class="check"><input type="checkbox" checked={collection.titleIds.includes(title.id)} onchange={event=>toggleTitle(title.id,event.currentTarget.checked)}/>{title.name}</label>{/each}{#if more}<button type="button" class="button subtle" disabled={busy} onclick={()=>operation(loadTitles)}>{text($locale,'Тағы тайтлдар','Ещё тайтлы')}</button>{/if}</fieldset>
<button class="button primary" disabled={busy}>{text($locale,'Жинақты сақтау','Сохранить подборку')}</button>{#if collection.id}<button class="button subtle" type="button" disabled={busy} onclick={()=>operation(async()=>{await gateway.deleteTaxon('collections',collection.id);collection=emptyCollection();await reload();})}>{text($locale,'Жинақты жою','Удалить подборку')}</button>{/if}
</form></section></div>
<style>
 form{display:grid;gap:12px}.panel{min-width:0}fieldset{border:1px solid var(--line);padding:12px;max-height:380px;overflow:auto}.check{display:flex;align-items:center;gap:10px;min-height:44px;overflow-wrap:anywhere}.check input{width:20px;height:20px;min-height:20px;flex:none}ol{padding-left:24px}li{display:flex;gap:6px;align-items:center;min-width:0;margin:6px 0}li span{flex:1;overflow-wrap:anywhere}li button{min-width:44px;min-height:44px;background:var(--panel);border:1px solid var(--line);border-radius:6px;color:var(--text)}
</style>
