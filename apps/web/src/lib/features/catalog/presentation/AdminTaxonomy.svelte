<script lang="ts">
 import ImportCandidates from './ImportCandidates.svelte';
 import type {MetadataImportsGateway} from '../application/metadata-import';
 import UnsavedChanges from '$lib/shared/presentation/ui/UnsavedChanges.svelte';
 import {fieldValidation} from '$lib/shared/presentation/form-validation';
 import {searchTaxonomy} from '../domain/taxonomy-search';
 import {onMount,tick} from 'svelte';
 import type {CatalogManagementGateway} from '../application/management';
 import type {Genre,Collection,Taxonomy,Summary} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 let {gateway,imports,actor=''}:{gateway:CatalogManagementGateway;imports?:MetadataImportsGateway;actor?:string}=$props();
 const {locale}=useLocale();
 let taxonomy=$state<Taxonomy>({genres:[],categories:[],collections:[]});
 let taxonQuery=$state(''),dictionary=$state<'genres'|'categories'|'collections'>('genres');let dictionaryEditor=$state<HTMLElement>();
 async function focusEditor(){await tick();dictionaryEditor?.querySelector<HTMLInputElement>('input:not([type=search])')?.focus();}
 function chooseDictionary(value:typeof dictionary){if(busy||!guard.confirmDiscard())return;resetTaxon();resetCollection();dictionary=value;if(value!=='collections')kind=value;taxonQuery='';}
 function taxonName(item:Genre){return $locale==='ru'&&item.nameRu?item.nameRu:item.name;}
 function taxonGroup(item:Genre){const group=taxonomy.categories.find(group=>group.slug===item.categorySlug);return group?taxonName(group):text($locale,'Топсыз','Без группы');}
 let kind=$state<'genres'|'categories'>('genres'), taxon=$state<Genre>({id:'',name:'',slug:''});
 const filteredTaxa=$derived(searchTaxonomy(taxonomy[kind],taxonQuery));
 const emptyCollection=():Collection=>({id:'',name:'',slug:'',description:'',position:0,visible:false,titleIds:[]});
 let collection=$state<Collection>(emptyCollection()), titles=$state<Summary[]>([]),more=$state(false),busy=$state(false),failure=$state(''),notice=$state('');
 let guard:UnsavedChanges;
 let taxonSaved=$state(JSON.stringify({id:'',name:'',slug:''})),collectionSaved=$state(JSON.stringify(emptyCollection()));
 const dirty=$derived(JSON.stringify(taxon)!==taxonSaved||JSON.stringify(collection)!==collectionSaved);
 function resetTaxon(){taxon={id:'',name:'',slug:''};taxonSaved=JSON.stringify(taxon);}
 function resetCollection(){collection=emptyCollection();collectionSaved=JSON.stringify(collection);}
 async function reload(){taxonomy=await gateway.taxonomy();}
 async function operation(action:()=>Promise<void>,saved=true){if(busy)return;busy=true;failure='';notice='';try{await action();if(saved)notice='saved';}catch(error){failure=failureCode(error);}finally{busy=false;}}
 async function loadTitles(){const next=await gateway.titles(titles.length);titles=[...titles,...next];more=next.length===100;}
 onMount(()=>{void operation(async()=>{await reload();await loadTitles();},false);});
 function chooseTaxon(id:string){if(!guard.confirmDiscard())return;taxon={...(taxonomy[kind].find(item=>item.id===id)||{id:'',name:'',slug:''})};taxonSaved=JSON.stringify(taxon);void focusEditor();}
 function chooseCollection(id:string){if(!guard.confirmDiscard())return;const item=taxonomy.collections.find(item=>item.id===id);collection=item?{...item,titleIds:[...item.titleIds]}:emptyCollection();collectionSaved=JSON.stringify(collection);void focusEditor();}
 function toggleTitle(id:string,checked:boolean){collection.titleIds=checked?[...collection.titleIds,id]:collection.titleIds.filter(value=>value!==id);}
 function move(index:number,direction:number){const next=[...collection.titleIds],target=index+direction;if(target<0||target>=next.length)return;[next[index],next[target]]=[next[target],next[index]];collection.titleIds=next;}
</script>
<UnsavedChanges bind:this={guard} {dirty}/>
<svelte:head><title>{text($locale,'Жанрлар мен жинақтар','Жанры и подборки')} — Qmanga</title></svelte:head>
<h1>{text($locale,'Жанрлар, санаттар және жинақтар','Жанры, категории и подборки')}</h1>
<p class="muted">{text($locale,'Жанр — сүзгі. Санат — мазмұн белгісі. Жинақ — редакция таңдаған тайтлдар. Жас санаты тайтлда бөлек көрсетіледі.','Жанр служит фильтром, категория отмечает содержание, подборка объединяет выбранные редакцией тайтлы. Возраст указывается отдельно в тайтле.')}</p>
{#if failure}<p role="alert">{noticeText($locale,failure)}</p>{/if}{#if notice}<p role="status">{noticeText($locale,notice)}</p>{/if}
<nav class="dictionary-tabs" aria-label={text($locale,'Анықтамалық түрі','Вид справочника')}>{#each [{id:'genres',kk:'Жанрлар',ru:'Жанры'},{id:'categories',kk:'Санаттар',ru:'Категории'},{id:'collections',kk:'Жинақтар',ru:'Подборки'}] as tab}<button class:chosen={dictionary===tab.id} aria-pressed={dictionary===tab.id} disabled={busy} onclick={()=>chooseDictionary(tab.id as typeof dictionary)}>{text($locale,tab.kk,tab.ru)} <span>{taxonomy[tab.id as typeof dictionary].length}</span></button>{/each}</nav>
<div class="dictionary-workbench"><section class="dictionary-list" aria-label={text($locale,'Анықтамалық жазбалары','Записи справочника')}>
{#if dictionary!=='collections'}<label>{text($locale,'Тег іздеу · қазақша немесе орысша','Поиск тега · на казахском или русском')}<input type="search" bind:value={taxonQuery}/></label><div class="dictionary-columns"><span>{text($locale,'Атауы','Название')}</span><span>{text($locale,'Сілтеме атауы','Имя в ссылке')}</span></div>{#each filteredTaxa as item}<button class="dictionary-row" class:chosen={taxon.id===item.id} aria-pressed={taxon.id===item.id} disabled={busy} onclick={()=>chooseTaxon(item.id)}><span><strong>{taxonName(item)}</strong>{#if item.nameRu&&item.nameRu!==item.name}<small>{item.name}</small>{/if}</span><span>{item.slug}</span></button>{:else}<p class="empty-state">{text($locale,'Тег табылмады','Теги не найдены')}</p>{/each}<button class="button subtle dictionary-create" disabled={busy} onclick={()=>chooseTaxon('')}>{text($locale,'Жаңасын қосу','Добавить запись')}</button>
{:else}<div class="dictionary-columns"><span>{text($locale,'Атауы','Название')}</span><span>{text($locale,'Тайтлдар','Тайтлы')}</span></div>{#each taxonomy.collections as item}<button class="dictionary-row" class:chosen={collection.id===item.id} aria-pressed={collection.id===item.id} disabled={busy} onclick={()=>chooseCollection(item.id)}><strong>{item.name}</strong><span>{item.titleIds.length}</span></button>{:else}<p class="empty-state">{text($locale,'Әзірге жинақ жоқ','Подборок пока нет')}</p>{/each}<button class="button subtle dictionary-create" disabled={busy} onclick={()=>chooseCollection('')}>{text($locale,'Жаңа жинақ','Создать подборку')}</button>{/if}</section>
<section class="panel dictionary-editor" bind:this={dictionaryEditor}>{#if dictionary!=='collections'}<h2>{taxon.id?text($locale,'Жазбаны өңдеу','Редактирование записи'):text($locale,'Жаңа жазба','Новая запись')}</h2>
<form use:fieldValidation={{locale:$locale}} onsubmit={event=>{event.preventDefault();void operation(async()=>{await gateway.saveTaxon(kind,taxon);resetTaxon();await reload();});}}>
<label>{text($locale,'Атауы','Название')}<input bind:value={taxon.name} required maxlength="80"/></label><label>{text($locale,'Сілтеме атауы','Имя в ссылке')}<input bind:value={taxon.slug} required maxlength="64" pattern="[a-z0-9]+(-[a-z0-9]+)*"/></label>
<button class="button primary" disabled={busy}>{text($locale,'Сақтау','Сохранить')}</button>
{#if taxon.id}<button class="button subtle" type="button" disabled={busy} onclick={()=>operation(async()=>{await gateway.deleteTaxon(kind,taxon.id);resetTaxon();await reload();})}>{text($locale,'Жою','Удалить')}</button>{/if}
</form>{:else}<h2>{text($locale,'Редакциялық жинақтар','Редакционные подборки')}</h2>
<form use:fieldValidation={{locale:$locale}} onsubmit={event=>{event.preventDefault();void operation(async()=>{await gateway.saveCollection(collection);resetCollection();await reload();});}}>
<label>{text($locale,'Таңдау','Выбрать')}<select value={collection.id} onchange={event=>{chooseCollection(event.currentTarget.value);event.currentTarget.value=collection.id;}}><option value="">{text($locale,'Жаңа жинақ','Создать подборку')}</option>{#each taxonomy.collections as item}<option value={item.id}>{item.name}</option>{/each}</select></label>
<label>{text($locale,'Атауы','Название')}<input bind:value={collection.name} required maxlength="80"/></label><label>{text($locale,'Сілтеме атауы','Имя в ссылке')}<input bind:value={collection.slug} required maxlength="64" pattern="[a-z0-9]+(-[a-z0-9]+)*"/></label>
<label>{text($locale,'Сипаттама','Описание')}<textarea bind:value={collection.description} maxlength="4000" rows="4"></textarea></label><label>{text($locale,'Реті','Порядок')}<input type="number" bind:value={collection.position} min="0" max="1000000" required/></label>
<label class="check"><input type="checkbox" bind:checked={collection.visible}/>{text($locale,'Каталогта көрінеді','Видна в каталоге')}</label>
<h3>{text($locale,'Тайтлдардың реті','Порядок тайтлов')}</h3>
<ol>{#each collection.titleIds as id,index}<li><span>{titles.find(item=>item.id===id)?.name||id}</span><button type="button" disabled={index===0} onclick={()=>move(index,-1)} aria-label={text($locale,'Жоғары жылжыту','Поднять')}>↑</button><button type="button" disabled={index===collection.titleIds.length-1} onclick={()=>move(index,1)} aria-label={text($locale,'Төмен жылжыту','Опустить')}>↓</button><button type="button" onclick={()=>toggleTitle(id,false)} aria-label={text($locale,'Жинақтан алып тастау','Убрать из подборки')}>×</button></li>{/each}</ol>
<fieldset><legend>{text($locale,'Тайтлдарды таңдау','Выбрать тайтлы')}</legend>{#each titles as title}<label class="check"><input type="checkbox" checked={collection.titleIds.includes(title.id)} onchange={event=>toggleTitle(title.id,event.currentTarget.checked)}/>{title.name}</label>{/each}{#if more}<button type="button" class="button subtle" disabled={busy} onclick={()=>operation(loadTitles,false)}>{text($locale,'Тағы тайтлдар','Ещё тайтлы')}</button>{/if}</fieldset>
<button class="button primary" disabled={busy}>{text($locale,'Жинақты сақтау','Сохранить подборку')}</button>{#if collection.id}<button class="button subtle" type="button" disabled={busy} onclick={()=>operation(async()=>{await gateway.deleteTaxon('collections',collection.id);resetCollection();await reload();})}>{text($locale,'Жинақты жою','Удалить подборку')}</button>{/if}
</form>{/if}</section></div>
{#if imports&&actor}<ImportCandidates gateway={imports} {actor} {taxonomy}/>{/if}
<style>
 .dictionary-tabs{display:flex;gap:8px;flex-wrap:wrap;margin:24px 0}.dictionary-tabs button{min-height:44px;padding:10px 16px;background:none;color:var(--muted);border:1px solid transparent;border-radius:8px;font-size:13px}.dictionary-tabs button.chosen{background:var(--raised);color:var(--accent)}.dictionary-tabs span{font-size:11px;margin-left:6px}.dictionary-workbench{display:grid;gap:24px;align-items:start}.dictionary-list{min-width:0;max-height:70dvh;overflow:auto;border:1px solid var(--line);border-radius:12px}.dictionary-list>label{display:grid;gap:8px;font-size:12px;color:var(--muted);padding:16px}.dictionary-row,.dictionary-columns{display:grid;grid-template-columns:minmax(0,1.4fr) minmax(0,1fr);gap:16px;padding:16px 18px;text-align:start;overflow-wrap:anywhere}.dictionary-columns{background:var(--raised);color:var(--muted);font-size:11px;text-transform:uppercase}.dictionary-row{width:100%;min-height:64px;background:none;border:0;border-bottom:1px solid var(--line);color:var(--text);align-items:center;font-size:13px}.dictionary-row.chosen{background:var(--raised)}.dictionary-row strong{font-weight:500}.dictionary-row small{display:block;color:var(--muted);margin-top:6px}.dictionary-row>span:last-child{color:var(--muted);font-size:12px}.dictionary-create{margin:16px}.dictionary-editor{scroll-margin-top:88px}.dictionary-editor h2{font-size:18px}.dictionary-list .empty-state{padding:24px}@media(min-width:1100px){.dictionary-workbench{grid-template-columns:minmax(0,1.6fr) minmax(300px,1fr)}.dictionary-editor{position:sticky;top:88px;padding:24px}}
 form{display:grid;gap:12px}.panel{min-width:0}fieldset{border:1px solid var(--line);padding:12px;max-height:380px;overflow:auto}.check{display:flex;align-items:center;gap:10px;min-height:44px;overflow-wrap:anywhere}.check input{width:20px;height:20px;min-height:20px;flex:none}ol{padding-left:24px}li{display:flex;gap:6px;align-items:center;min-width:0;margin:6px 0}li span{flex:1;overflow-wrap:anywhere}li button{min-width:44px;min-height:44px;background:var(--panel);border:1px solid var(--line);border-radius:6px;color:var(--text)}
</style>
