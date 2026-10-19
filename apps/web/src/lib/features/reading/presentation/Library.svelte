<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import LibrarySyncStatus from './LibrarySyncStatus.svelte';
 import type {LibrarySync} from '../application/library-sync';
 import LibraryRow from './LibraryRow.svelte';
 import ProgressPosition from './ProgressPosition.svelte';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import {onMount,untrack} from 'svelte';
 import type {LibraryEntry,LocalProgress,Progress,ProgressStorage,ReadingGateway} from '../application/progress';
 import {importGuest as transferGuest} from '../application/guest-import';
 import {LibraryDirectory} from '../application/library-directory';

 let {readingGateway,progressStorage,librarySync}:{readingGateway:ReadingGateway;progressStorage:ProgressStorage;librarySync?:LibrarySync}=$props();
 let guests=$state<LocalProgress[]>([]),account=$state<string|null>(null),message=$state(''),filter=$state('all'),conflict=$state<{guest:LocalProgress;cloud:Progress}|null>(null),valid=true;
 let directoryVersion=$state(0);
 const directory=new LibraryDirectory(untrack(()=>readingGateway),()=>directoryVersion++);
 const entries=$derived.by(()=>{directoryVersion;return directory.items;});
 const cursor=$derived.by(()=>{directoryVersion;return directory.cursor;});
 const hasMore=$derived.by(()=>{directoryVersion;return directory.hasMore;});
 const total=$derived.by(()=>{directoryVersion;return directory.total;});
 const partial=$derived.by(()=>{directoryVersion;return directory.partial;});
 const busy=$derived.by(()=>{directoryVersion;return directory.busy;});
 const directoryError=$derived.by(()=>{directoryVersion;return directory.error;});

 const statusNames=$derived<Record<string,string>>({reading:copy($locale,"Оқып жүрмін"),planned:copy($locale,"Жоспарда"),completed:copy($locale,"Оқып болдым"),paused:copy($locale,"Үзілісте"),dropped:copy($locale,"Тоқтаттым")});


 let savingTitle=$state('');

 onMount(()=>{
  const stop=progressStorage.watchAccountChange(()=>{valid=false;directory.clear();account=null;conflict=null;message="account_changed";});
  void init();
  return ()=>{valid=false;directory.clear();stop();};
 });

 async function init(){
  try{
   const current=account||await readingGateway.account();
   const guestRows=await progressStorage.list('guest');
   if(!valid)return;
   guests=guestRows.sort((left,right)=>(right.updatedAt||0)-(left.updatedAt||0));
   account=current;
   if(current)await directory.search(current,'',filter);
  }catch(failure){
   if(valid)message=failureCode(failure);
  }
 }

 async function handleFilterChange(){
  if(!valid||!account)return;
  message='';
  await directory.search(account,'',filter);
 }

 async function update(item:LibraryEntry){
  if(!valid||!account||directory.busy||savingTitle)return;
  const scope=account;
  savingTitle=item.titleId;
  message='saving';
  try{
   await readingGateway.save(scope,{...item});
   const local=await librarySync?.state(scope,item.titleId);
   if(!valid||account!==scope)return;
   if(local?.pending){
    directory.applyPendingUpdate({...item,...local.selection});
    message='';
   }else{
    await directory.refresh(scope);
    if(valid&&account===scope)message='saved';
   }
  }catch(failure){
   if(valid&&account===scope)message=failureCode(failure);
  }finally{
   savingTitle='';
  }
 }

 async function remove(item:LibraryEntry){
  if(!valid||!account||directory.busy||savingTitle)return;
  const scope=account;
  savingTitle=item.titleId;
  message='saving';
  try{
   await readingGateway.remove(scope,item.titleId);
   const local=await librarySync?.state(scope,item.titleId);
   if(!valid||account!==scope)return;
   if(local?.pending){
    directory.applyPendingUpdate({...item,...local.selection});
    message='';
   }else{
    await directory.refresh(scope);
    if(valid&&account===scope)message='saved';
   }
  }catch(failure){
   if(valid&&account===scope)message=failureCode(failure);
  }finally{
   savingTitle='';
  }
 }
 async function importGuest(guest:LocalProgress,baseVersion?:number){
  if(!valid||!account)return;
  try{
   const cloud=await transferGuest(readingGateway,progressStorage,account,guest,baseVersion);
   if(cloud){conflict={guest,cloud};message="progress_conflict";return;}
   conflict=null;
   message="guest_imported";
   await directory.refresh(account);
  }catch{
   message="import_failed";
  }
 }
</script>
<svelte:head><title>{copy($locale,"Кітапхана — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={copy($locale,"Кітапхана")}/>
{#if directoryError}
 <div class="error-notice">
  <p role="alert" class="form-message">{noticeText($locale,directoryError)}</p>
  <button class="button subtle" disabled={busy||!!savingTitle} onclick={()=>{if(valid&&account){message='';void directory.retry(account);}}}>{text($locale,'Қайталау','Повторить')}</button>
 </div>
{:else if message}
 <p role="status" class="form-message">{noticeText($locale,message)}</p>
{/if}
{#if account&&librarySync}<LibrarySyncStatus sync={librarySync} {account} onresolved={()=>{if(valid&&account)void directory.refresh(account);}}/>{/if}
{#if account}
 <nav class="library-filters" aria-label={copy($locale,"Тізім")}>{#each [['all',copy($locale,"Барлығы")],['favorites',copy($locale,"Таңдаулы")],...Object.entries(statusNames)] as [value,label]}<button aria-pressed={filter===value} disabled={busy||!!savingTitle} onclick={()=>{filter=value;void handleFilterChange();}}>{label}</button>{/each}</nav>
 {#if partial}<p class="muted">{text($locale,'Жергілікті көшірме немесе синхрондалмаған өзгерістер. Толық сан әзірге белгісіз.','Локальная копия или несинхронизированные изменения. Полное число пока неизвестно.')}</p>{:else if total!==undefined}<p class="muted">{text($locale,'Табылды: ','Найдено: ')}{total}</p>{/if}
 <div class="library-grid">{#each entries as item (item.titleId)}<LibraryRow coverURL={item.coverURL} name={item.name} href={`/titles/${item.slug}`} detail={text($locale,'Оқылмаған тараулар:','Непрочитанных глав:')+' '+item.unseen} releaseId={item.progress.releaseId} continueLabel={copy($locale,"Жалғастыру")}>{#if item.progress.releaseId}<ProgressPosition progress={item.progress}/>{/if}<div class="entry-controls"><select aria-label={item.name+': '+text($locale,'мәртебесі','статус')} bind:value={item.status} disabled={busy||!!savingTitle} onchange={()=>update(item)}>{#each Object.entries(statusNames) as [value,label]}<option {value}>{label}</option>{/each}</select><label><input type="checkbox" bind:checked={item.favorite} disabled={busy||!!savingTitle} onchange={()=>update(item)}/>{copy($locale,"Таңдаулы")}</label><button class="button subtle" disabled={busy||!!savingTitle} onclick={()=>remove(item)}>{copy($locale,"Тізімнен алу")}</button>{#if savingTitle===item.titleId}<span class="muted">{noticeText($locale,"saving")}</span>{/if}</div></LibraryRow>{:else}{#if !busy&&!message&&!directoryError}<p class="empty-state">{copy($locale,"Бұл тізімде тайтл жоқ. Тайтл бетінде кітапханаға қосыңыз.")}</p>{/if}{/each}</div>
 {#if hasMore}<button class="button" disabled={busy||!!savingTitle} onclick={()=>directory.more(account!)}>{copy($locale,"Тағы жүктеу")}</button>{/if}
{:else if !busy}<p class="muted">{copy($locale,"Қонақтағы оқу орны осы құрылғыда сақталады. Құрылғылар арасында сақтау үшін ")}<a class="auth-link" href="/login?returnTo=%2Flibrary">{copy($locale,"кіріңіз")}</a>.</p>{/if}
{#if busy}<p role="status">{copy($locale,"Жүктелуде…")}</p>{/if}

{#if guests.length}<section><h2>{copy($locale,"Қонақ ретінде оқылған")}</h2><p class="muted">{copy($locale,"Есептікжазбаға автоматты қосылмайды. Импортты өзіңіз таңдаңыз.")}</p>{#each guests as guest}<article><a class="title" href={`/titles/${guest.slug}`}>{guest.name}</a><ProgressPosition progress={guest.progress} chapterLabel={guest.chapterLabel} pagePosition={guest.pagePosition} updatedAt={guest.updatedAt}/><div class="entry-controls"><a class="button subtle" href={`/read/${guest.progress.releaseId}`}>{copy($locale,"Оқу")}</a>{#if account}<button class="button subtle" disabled={busy||!!savingTitle} onclick={()=>importGuest(guest)}>{copy($locale,"Осы орынды импорттау")}</button>{/if}</div></article>{/each}</section>{/if}
{#if conflict}<section class="conflict"><h2>{copy($locale,"Оқу орны әртүрлі")}</h2><p>{conflict.guest.name}: {text($locale,'бұлтта да оқу орны бар. Импорт оны ауыстырады.','в облаке тоже есть место чтения. Импорт заменит его.')}</p><p>{text($locale,'Осы құрылғыда','На этом устройстве')}</p><ProgressPosition progress={conflict.guest.progress} chapterLabel={conflict.guest.chapterLabel} pagePosition={conflict.guest.pagePosition} updatedAt={conflict.guest.updatedAt}/><p>{text($locale,'Бұлтта','В облаке')}</p><ProgressPosition progress={conflict.cloud}/><div class="entry-controls"><button class="button primary" onclick={()=>importGuest(conflict!.guest,conflict!.cloud.version)} disabled={busy||!!savingTitle}>{copy($locale,"Қонақтағы орынды сақтау")}</button><button class="button subtle" onclick={()=>conflict=null}>{copy($locale,"Бұлттағыны қалдыру")}</button></div></section>{/if}
<style>.library-filters{display:flex;gap:4px;overflow-x:auto;margin:24px 0 16px;padding:4px;border:1px solid var(--line);border-radius:12px;background:var(--panel)}.library-filters button{flex:none;min-height:44px;padding:8px 12px;border:0;border-radius:8px;background:transparent;color:var(--muted);font-size:14px;cursor:pointer}.library-filters button[aria-pressed="true"]{background:var(--raised);color:var(--accent);font-weight:600}.library-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:20px 10px;margin-bottom:24px}.library-grid>.empty-state{grid-column:1/-1}@media(min-width:600px){.library-grid{grid-template-columns:repeat(4,minmax(0,1fr));gap:24px 16px}}@media(min-width:1000px){.library-grid{grid-template-columns:repeat(6,minmax(0,1fr))}}article{border-bottom:1px solid var(--line);padding:20px 0}.title{font-size:18px;font-weight:600;display:inline-flex;align-items:center;min-height:44px}.entry-controls{display:flex;align-items:center;flex-wrap:wrap;gap:12px;margin-top:12px}.entry-controls select{max-width:240px}.entry-controls label{display:flex;align-items:center;gap:8px;min-height:44px}.entry-controls input{width:20px;height:20px;min-height:20px}section{margin-top:36px}h2{font-size:21px}.conflict{border:1px solid var(--line);padding:20px}</style>
