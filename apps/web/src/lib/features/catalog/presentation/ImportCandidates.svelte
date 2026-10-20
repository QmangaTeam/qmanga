<script lang="ts">
 import {onMount} from 'svelte';
 import type {MetadataImportsGateway,ImportCandidate} from '../application/metadata-import';
 import type {Taxonomy} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {gateway,actor,taxonomy}:{gateway:MetadataImportsGateway;actor:string;taxonomy:Taxonomy}=$props();
 const {locale}=useLocale();let items=$state<ImportCandidate[]>([]),busy=$state(false),error=$state(false),targets=$state<Record<string,string>>({});
 async function load(){busy=true;error=false;try{items=await gateway.candidates(actor)}catch{error=true}finally{busy=false}}
 async function resolve(item:ImportCandidate){const value=targets[item.source+':'+item.id];if(!value)return;const [kind,target]=value.split(':');busy=true;error=false;try{await gateway.resolve(actor,item,kind as 'genres'|'categories',target);items=items.filter(x=>x.source!==item.source||x.id!==item.id)}catch{error=true}finally{busy=false}}
 onMount(()=>{void load()});
</script>
<section class="panel candidates"><h2>{text($locale,'Сыртқы тегтерді сәйкестендіру','Сопоставление внешних тегов')}</h2><p>{text($locale,'Жаңа сүзгілер автоматты жасалмайды. Қажет болса, алдымен жоғарыдағы формада тег жасаңыз, содан кейін сәйкестікті таңдаңыз.','Новые фильтры автоматически не создаются. При необходимости сначала создайте тег в форме выше, затем выберите соответствие.')}</p>
<button class="button subtle" type="button" disabled={busy} onclick={load}>{text($locale,'Жаңарту','Обновить')}</button>
{#if error}<p role="alert">{text($locale,'Деректер сақталмады. Қайталап көріңіз.','Не удалось сохранить данные. Повторите попытку.')}</p>{/if}
{#if busy}<p role="status">{text($locale,'Жүктелуде…','Загрузка…')}</p>{/if}
{#each items as item}<div class="candidate"><p><strong>{item.name}</strong> · {item.source} #{item.id} · {item.status==='ambiguous'?text($locale,'Бірнеше сәйкестік','Несколько совпадений'):text($locale,'Сәйкестік жоқ','Нет соответствия')}</p><label>{text($locale,'Каталогтағы тег','Тег каталога')}<select bind:value={targets[item.source+':'+item.id]} disabled={busy}><option value="">{text($locale,'Таңдаңыз','Выберите')}</option>{#each ['genres','categories'] as kind}<optgroup label={kind==='genres'?text($locale,'Жанрлар','Жанры'):text($locale,'Санаттар','Категории')}>{#each taxonomy[kind as 'genres'|'categories'] as tag}<option value={kind+':'+tag.id}>{tag.name} · {tag.slug}</option>{/each}</optgroup>{/each}</select></label><button class="button primary" type="button" disabled={busy||!targets[item.source+':'+item.id]} onclick={()=>resolve(item)}>{text($locale,'Сәйкестікті сақтау','Сохранить соответствие')}</button></div>{:else}{#if !busy&&!error}<p>{text($locale,'Келісуді қажет ететін тегтер жоқ.','Нет тегов для согласования.')}</p>{/if}{/each}</section>
<style>.candidates{margin-top:24px;overflow-wrap:anywhere}.candidate{border-top:1px solid var(--line);margin-top:16px;padding-top:8px;display:grid;gap:12px}label{display:grid;gap:8px;min-width:0}select{min-width:0;max-width:100%}.candidate button{justify-self:start}</style>
