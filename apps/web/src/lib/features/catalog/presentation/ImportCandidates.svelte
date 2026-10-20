<script lang="ts">
 import {onDestroy,untrack} from 'svelte';
 import type {MetadataImportsGateway,ImportCandidate} from '../application/metadata-import';
 import type {Taxonomy} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {gateway,actor,taxonomy}:{gateway:MetadataImportsGateway;actor:string;taxonomy:Taxonomy}=$props();
 const {locale}=useLocale();let items=$state<ImportCandidate[]>([]),busy=$state(false),error=$state(false),active=$state(''),target=$state('');let generation=0;
 const key=(item:ImportCandidate)=>item.source+':'+item.id;
 $effect(()=>{const account=actor;untrack(()=>{generation++;items=[];active='';target='';error=false;busy=false;if(account)void load()})});
 onDestroy(()=>{generation++});
 async function load(){const run=++generation,account=actor;busy=true;error=false;try{const result=await gateway.candidates(account);if(run===generation)items=result}catch{if(run===generation)error=true}finally{if(run===generation)busy=false}}
 async function resolve(item:ImportCandidate){if(!target||busy)return;const run=++generation,account=actor,[kind,id]=target.split(':');busy=true;error=false;try{await gateway.resolve(account,item,kind as 'genres'|'categories',id);if(run===generation){items=items.filter(x=>key(x)!==key(item));active='';target=''}}catch{if(run===generation)error=true}finally{if(run===generation)busy=false}}
</script>
<section class="panel candidates"><h2>{text($locale,'Сыртқы тегтерді сәйкестендіру','Сопоставление внешних тегов')}</h2><p>{text($locale,'Жаңа сүзгілер автоматты жасалмайды. Қажет болса, алдымен жоғарыдағы формада тег жасаңыз, содан кейін сәйкестікті таңдаңыз.','Новые фильтры автоматически не создаются. При необходимости сначала создайте тег в форме выше, затем выберите соответствие.')}</p>
<button class="button subtle" type="button" disabled={busy||!actor} onclick={load}>{text($locale,'Жаңарту','Обновить')}</button>
{#if error}<p role="alert">{text($locale,'Деректер сақталмады. Қайталап көріңіз.','Не удалось сохранить данные. Повторите попытку.')}</p>{/if}
{#if busy}<p role="status">{text($locale,'Жүктелуде…','Загрузка…')}</p>{/if}
{#each items as item}<div class="candidate"><p><strong>{item.name}</strong> · {item.source} #{item.id} · {item.status==='ambiguous'?text($locale,'Бірнеше сәйкестік','Несколько совпадений'):text($locale,'Сәйкестік жоқ','Нет соответствия')}</p>
{#if active===key(item)}<label>{text($locale,'Каталогтағы тег','Тег каталога')}<select bind:value={target} disabled={busy}><option value="">{text($locale,'Таңдаңыз','Выберите')}</option>{#each ['genres','categories'] as kind}<optgroup label={kind==='genres'?text($locale,'Жанрлар','Жанры'):text($locale,'Санаттар','Категории')}>{#each taxonomy[kind as 'genres'|'categories'] as tag}<option value={kind+':'+tag.id}>{tag.name} · {tag.slug}</option>{/each}</optgroup>{/each}</select></label><button class="button primary" type="button" disabled={busy||!target} onclick={()=>resolve(item)}>{text($locale,'Сәйкестікті сақтау','Сохранить соответствие')}</button>
{:else}<button class="button subtle" type="button" disabled={busy} onclick={()=>{active=key(item);target=''}}>{text($locale,'Сәйкестендіру','Сопоставить')}</button>{/if}</div>{:else}{#if !busy&&!error}<p>{text($locale,'Келісуді қажет ететін тегтер жоқ.','Нет тегов для согласования.')}</p>{/if}{/each}</section>
<style>.candidates{margin-top:24px;overflow-wrap:anywhere}.candidate{border-top:1px solid var(--line);margin-top:16px;padding-top:8px;display:grid;gap:12px}label{display:grid;gap:8px;min-width:0}select{min-width:0;max-width:100%}.candidate button{justify-self:start}</style>
