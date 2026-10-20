<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import {onMount} from 'svelte';
 import type {Downloads,Download} from '../application/downloads';
 let {downloads}:{downloads:Downloads}=$props();
 let entries=$state<Download[]>([]),message=$state(''),busy=$state(true),saving=$state('');
 let selected=$state<string[]>([]),confirming=$state(false),cleaning=$state(false),available=$state<number|null>(null),cleanupResult=$state('');
 const total=$derived(entries.reduce((sum,item)=>sum+item.bytes,0));
 async function cleanup(){if(cleaning||saving)return;const ids=[...selected];cleaning=true;confirming=false;try{const result=await downloads.removeSelected(ids);selected=result.failed;cleanupResult=text($locale,`Өшірілді: ${result.removed.length}; қате: ${result.failed.length}.`,`Удалено: ${result.removed.length}; ошибок: ${result.failed.length}.`);await load();}catch{message='remove_copy_failed';}finally{cleaning=false;}}
 let controller:AbortController|undefined;
 async function load(){busy=true;message='';try{entries=await downloads.list();available=await downloads.available().catch(()=>null);selected=selected.filter(id=>entries.some(item=>item.id===id));}catch{entries=[];message="storage_unavailable";}finally{busy=false;}}
 onMount(()=>{void load();return ()=>controller?.abort();});
 async function verify(item:Download){try{await downloads.reconcile($state.snapshot(item));await load();}catch{message="verify_failed";}}
 async function remove(id:string){try{await downloads.remove(id);await load();}catch{message="remove_copy_failed";}}
 async function resume(item:Download){
  if(saving)return;
  saving=item.id;message='';controller=new AbortController();
  try{await downloads.save($state.snapshot(item.manifest),item.tier,controller.signal,value=>{entries=entries.map(entry=>entry.id===item.id?value:entry);});}
  catch(failure){message=failureCode(failure);}
  finally{saving='';controller=undefined;}
 }
</script>
<svelte:head><title>{copy($locale,"Сақталған тараулар — Qmanga")}</title></svelte:head>
<h1>{copy($locale,"Сақталған тараулар")}</h1><p class="muted">{copy($locale,"Осы браузердегі көшірмелер. Браузер деректерді тазаласа, қайта жүктеу қажет.")}</p>
{#if !busy&&entries.length}<section class="storage-overview"><p>{text($locale,'Көшірмелер көлемі','Объём копий')}: {(total/1024/1024).toFixed(1)} МБ · {text($locale,'Браузердегі бос орын','Свободное место в браузере')}: {available===null?text($locale,'белгісіз','неизвестно'):(available/1024/1024).toFixed(1)+' МБ'}</p><p>{text($locale,'Тек таңдалған офлайн көшірмелер өшіріледі. Оқу барысы сақталады.','Удаляются только выбранные офлайн копии. Прогресс чтения сохраняется.')}</p><button class="button" disabled={!selected.length||!!saving||cleaning} onclick={()=>confirming=true}>{text($locale,'Таңдалғанын өшіру','Удалить выбранные')} ({selected.length})</button>{#if confirming}<p role="status">{text($locale,`${selected.length} көшірмені өшіру керек пе?`,`Удалить ${selected.length} копий?`)}</p><ul>{#each entries.filter(item=>selected.includes(item.id)) as item}<li>{item.manifest.titleName} · {item.manifest.chapterLabel} · {item.tier}</li>{/each}</ul><button class="button" onclick={cleanup}>{text($locale,'Өшіруді растау','Подтвердить удаление')}</button><button class="button" onclick={()=>confirming=false}>{text($locale,'Бас тарту','Отмена')}</button>{/if}</section>{/if}
{#if cleanupResult}<p role="status">{cleanupResult}</p>{/if}
{#if message}<p role="status" class="form-message">{noticeText($locale,message)}</p>{#if !entries.length}<button class="button" onclick={load} disabled={busy}>{text($locale,'Қайта көру','Повторить')}</button>{/if}{/if}
{#if busy}<p role="status">{text($locale,'Сақталған файлдар тексерілуде…','Проверяем сохранённые файлы…')}</p>{:else if !entries.length&&!message}<div class="empty-state"><h2>{copy($locale,"Сақталған тарау жоқ")}</h2><p>{copy($locale,"Читалкадағы «Баптаулар» арқылы тарауды құрылғыға сақтаңыз.")}</p><a href="/catalog" class="button">{copy($locale,"Каталогқа өту")}</a></div>{/if}
{#if !busy}{#each entries as item (item.id)}<article><label class="select-copy"><input type="checkbox" value={item.id} bind:group={selected} disabled={!!saving||cleaning||confirming}/>{text($locale,'Тазалауға таңдау','Выбрать для очистки')}</label><h2>{item.manifest.titleName}</h2><p>{text($locale,item.manifest.chapterLabel+'-тарау','Глава '+item.manifest.chapterLabel)} · {item.manifest.team.name}</p><p class="muted" aria-live={saving===item.id?'polite':'off'}>{item.status==='ready'?copy($locale,"Толық сақталған"):item.status==='partial'?copy($locale,"Толық емес"):copy($locale,"Сақтау қатесі")} · {item.tier} · {(item.bytes/1024/1024).toFixed(1)} / {(item.total/1024/1024).toFixed(1)} МБ</p>{#if item.status!=='ready'}<p class="muted">{text($locale,'Жетіспейтін файлдарды сақтау үшін желіге қосылыңыз.','Подключитесь к сети, чтобы скачать недостающие файлы.')}</p>{/if}<div>{#if item.status==='ready'}<a class="button primary" href={`/offline-reader?download=${encodeURIComponent(item.id)}`} data-sveltekit-reload>{copy($locale,"Желісіз оқу")}</a>{:else if saving===item.id}<button class="button" onclick={()=>controller?.abort()}>{text($locale,'Тоқтату','Остановить')}</button>{:else}<button class="button" disabled={!!saving} onclick={()=>resume(item)}>{copy($locale,"Жүктеуді жалғастыру")}</button>{/if}<button class="button subtle" disabled={!!saving} onclick={()=>verify(item)}>{copy($locale,"Тексеру")}</button><button class="button subtle" disabled={!!saving} onclick={()=>remove(item.id)}>{copy($locale,"Көшірмені өшіру")}</button></div></article>{/each}{/if}
<style>.select-copy{display:flex;align-items:center;gap:10px;min-height:44px}.select-copy input{width:20px;height:20px}.storage-overview{padding:16px;border:1px solid var(--line)}.storage-overview p{margin-bottom:12px}h1{font-size:26px;margin-bottom:8px}article{padding:16px;border:1px solid var(--line);border-radius:12px;background:var(--panel);margin-block:12px;min-width:0}article h2{font-size:18px}article p{font-size:14px;margin:8px 0}article>div{display:flex;flex-wrap:wrap;gap:8px;margin-top:16px}.empty-state{margin:24px 0}.empty-state h2{font-size:18px}.empty-state a{margin-top:20px}</style>
