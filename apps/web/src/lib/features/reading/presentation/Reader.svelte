<script lang="ts">
 import { onMount, tick, untrack } from 'svelte';
 import { anchorOffset, layout, visible, type ImageLoader, type Manifest, type ReaderGateway, type ReaderPlatform, type ReaderPreferences, type Tier, type Timer, type Viewport } from '../application/reader';
 import { ImageQueue } from '../application/scheduler';
 import {PageNavigation,readingDirection,pageStep,type PageRequest} from '../application/navigation';
 import { chooseQuality } from '../domain/policies';
 import type { CatalogGateway, Chapter } from '../../catalog/application/catalog';
 import type {Download,Downloads} from '../../offline/application/downloads';
 import {ReadingSession,type ProgressStorage,type ReadingGateway} from '../application/progress';
 import type {CommunityGateway} from '../../community/application/community';
 import Discussion from '../../community/presentation/Discussion.svelte';
 import ReportForm from '../../community/presentation/ReportForm.svelte';
 import {ThrottledSave} from '../application/throttled-save';
 import type {Position} from '../domain/policies';
 let {manifest,preview='',offlineTier,downloads,readingGateway,progressStorage,community,gateway,catalogue,platform,images,timer}:{manifest:Manifest;preview?:string;offlineTier?:Tier;downloads?:Downloads;readingGateway?:ReadingGateway;progressStorage?:ProgressStorage;community?:CommunityGateway;gateway:ReaderGateway;catalogue:CatalogGateway;platform:ReaderPlatform;images:(first?:HTMLImageElement)=>ImageLoader;timer:Timer}=$props();
 let preferences=$state<ReaderPreferences>({quality:'auto',mode:untrack(()=>manifest.kind)==='manga'?'pages':'strip',theme:'dark',fit:'width',maxWidth:1024,brightness:100,numbers:true,direction:'auto'});
 let tier=$state<Tier>(untrack(()=>offlineTier)||'standard'),pageIndex=$state(0),viewport=$state<Viewport>({width:360,height:640,top:0});
 let stage=$state<HTMLDivElement>(),firstImage=$state<HTMLImageElement>(),dialog=$state<HTMLDialogElement>(),shell=$state<HTMLDivElement>();
 let queue:ImageQueue|undefined,version=$state(0),mounted=$state(false),switching=$state(false),message=$state(''),panel=$state<''|'settings'|'chapters'>('');
 let closePanel:(()=>void)|undefined,saver:ThrottledSave<{manifest:Manifest;position:Position;completed:boolean}>|undefined;
 let chapters=$state<Chapter[]>([]),cursor=$state<string|null>(null),chapterBusy=$state(false),query=$state(''),reverse=$state(false),original=$state<{url:string;bytes:number}|null>(null);
 let download=$state<Download|null>(null),downloadBusy=$state(false),downloadController:AbortController|undefined;
 let session:ReadingSession|undefined,progressVersion=$state(0);
 const navigation=new PageNavigation(untrack(()=>manifest.pages.length));
 const progressMessage=$derived.by(()=>{progressVersion;return session?.message||'';});
 const conflict=$derived.by(()=>{progressVersion;return session?.row?.conflict;});
 const sheet=$derived(layout(manifest,tier,preferences.mode,pageIndex));
 const current=$derived(visible(sheet.items,viewport));
 const start=$derived(Math.max(0,current.start-1)),end=$derived(Math.min(sheet.items.length,current.end+2));
 const painted=$derived.by(()=>{version;return sheet.items.slice(start,end+1).map(item=>({item,state:queue?.entries.get(item.id)}));});
 const direction=$derived(readingDirection(preferences.mode,preferences.direction,manifest.direction));
 const leftStep=$derived(pageStep('left',direction)),rightStep=$derived(pageStep('right',direction));
 const canStep=(delta:number)=>pageIndex+delta>=0&&pageIndex+delta<manifest.pages.length;
 const shownChapters=$derived(chapters.filter(chapter=>(chapter.label+' '+chapter.name).toLocaleLowerCase('kk').includes(query.toLocaleLowerCase('kk'))).sort((left,right)=>reverse?left.order-right.order:right.order-left.order));
 const qualityNames={auto:'Автоматты',eco:'Үнемді',standard:'Қалыпты',high:'Жоғары'};
 $effect(()=>{if(!mounted||!queue||switching)return;const items=sheet.items;const wanted=[...items.slice(current.start,current.end+1),...items.slice(current.end+1,current.end+3),...items.slice(Math.max(0,current.start-1),current.start)];untrack(()=>queue?.want(wanted));});
 $effect(()=>{if(!mounted||!current.position.pageId||preview)return;const position=current.position,finished=current.pageIndex===manifest.pages.length-1&&viewport.top+viewport.height>=sheet.total*viewport.width-20;untrack(()=>saver?.offer({manifest,position:finished?{pageId:position.pageId,fraction:1}:position,completed:finished}));});
 $effect(()=>{if(mounted)platform.savePreferences(preferences);});
 onMount(()=>{
  queue=new ImageQueue(images(firstImage),timer,()=>version++);
  saver=new ThrottledSave(timer,async value=>{if(session)await session.record(value.manifest,value.position,value.completed);else if(!readingGateway)platform.savePosition(value.manifest.releaseId,value.position);});mounted=true;
  let observed=false;const stopObserve=platform.observe(stage,value=>{const position=current.position,resized=observed&&Math.abs(value.width-viewport.width)>.5;observed=true;viewport=value;if(preferences.mode==='strip'){navigation.observe(visible(sheet.items,value).pageIndex);pageIndex=navigation.index;}if(resized&&position.pageId)void tick().then(()=>platform.scroll(stage,anchorOffset(sheet.items,position,value.width)));}),stopVisibility=platform.visibility(hidden=>{queue?.pause(hidden);if(hidden)void saver?.flush().then(()=>session?.flush());}),stopKeyboard=platform.keyboard(action=>{if(panel)return;if(action==='menu')void openPanel('settings');else step(pageStep(action,direction));});
  preferences=platform.preferences(manifest.kind);
  if(offlineTier){preferences={...preferences,quality:offlineTier};tier=offlineTier;}
  const saved=preview||readingGateway?null:platform.loadPosition(manifest.releaseId);
  if(!preview&&readingGateway&&progressStorage){session=new ReadingSession(readingGateway,progressStorage,timer,()=>progressVersion++);void session.start(manifest,!!offlineTier).then(async position=>{if(position?.releaseId===manifest.releaseId&&manifest.pages.some(page=>page.id===position.pageId))await goPage(manifest.pages.findIndex(page=>page.id===position.pageId),position.fraction,false);await tick();saver?.offer({manifest,position:current.position,completed:false});});}
  void (async()=>{await changeQuality(preferences.quality);if(saved){const index=manifest.pages.findIndex(page=>page.id===saved.pageId);if(index>=0)await goPage(index,saved.fraction,false);}})();
  return ()=>{mounted=false;void saver?.flush().finally(()=>session?.close());downloadController?.abort();closePanel?.();stopObserve();stopVisibility();stopKeyboard();queue?.close();};
 });
 async function changeQuality(quality:ReaderPreferences['quality']) {
  preferences={...preferences,quality};const next=chooseQuality(quality,platform.connection()) as Tier;if(next===tier||!queue)return;
  const position=current.position,proposed=layout(manifest,next,preferences.mode,pageIndex),offset=anchorOffset(proposed.items,position,viewport.width),range=visible(proposed.items,{...viewport,top:offset}),target=proposed.items[range.start];if(!target)return;
  switching=true;message='';queue.want([target,...sheet.items.slice(current.start,current.end+1)]);
  try{await queue.ready(target.id);tier=next;await tick();platform.scroll(stage,offset);}catch(failure){message=failure instanceof Error?failure.message:'Сапаны ауыстыру мүмкін болмады';}finally{switching=false;}
 }
 async function mode(value:'pages'|'strip') {const position=current.position,index=current.pageIndex;preferences={...preferences,mode:value};await goPage(index,position.fraction,false);}
 function step(delta:number) {const request=navigation.move(delta);if(request)void navigate(request);else message=delta>0?'Тараудың соңы. Келесі тарауды төменнен таңдаңыз.':'Тараудың басы';}
 async function goPage(index:number,fraction=0,persist=true){const request=navigation.select(index);if(request)await navigate(request,fraction,persist);}
 async function navigate(request:PageRequest,fraction=0,persist=true){pageIndex=request.index;message='';const position={pageId:manifest.pages[request.index].id,fraction};await tick();if(!mounted||!navigation.current(request))return;platform.scroll(stage,anchorOffset(sheet.items,position,viewport.width));navigation.finish(request);if(persist&&!preview){saver?.offer({manifest,position,completed:false});void saver?.flush();}}
 async function openPanel(value:'settings'|'chapters') {panel=value;await tick();closePanel=platform.panel(dialog,()=>{panel='';closePanel=undefined;});if(value==='chapters'&&!chapters.length&&!preview)await loadChapters();}
 function dismiss() {const stop=closePanel;closePanel=undefined;panel='';stop?.();}
 async function loadChapters() {chapterBusy=true;try{const result=await catalogue.chapters(manifest.titleSlug,cursor||undefined);chapters=[...chapters,...result.items];cursor=result.nextCursor;}catch{message='Тараулар тізімін жүктеу мүмкін болмады';}finally{chapterBusy=false;}}
 async function getOriginal() {original=null;message='';try{original=await gateway.original(manifest.releaseId,manifest.pages[current.pageIndex].id);}catch{message='Түпнұсқа қолжетімсіз';}}
 const releaseFor=(chapter:Chapter)=>chapter.releases.find(release=>release.projectId===manifest.projectId);
 async function saveOffline(){if(!downloads)return;downloadBusy=true;downloadController=new AbortController();try{await downloads.save(manifest,tier,downloadController.signal,value=>download=value);}catch(failure){message=failure instanceof Error?failure.message:'Сақтау мүмкін болмады';}finally{downloadBusy=false;}}
 async function resolveProgress(local:boolean){const position=await session?.resolve(local);if(!local&&position?.releaseId===manifest.releaseId){const index=manifest.pages.findIndex(page=>page.id===position.pageId);if(index>=0)await goPage(index,position.fraction,false);}}
</script>

<svelte:head><title>{manifest.chapterLabel}-тарау · {manifest.titleName} — Qmanga</title><meta name="robots" content={preview?'noindex,nofollow':'index,follow'}/></svelte:head>
<div bind:this={shell} class="reader" class:light={preferences.theme==='light'} class:system={preferences.theme==='system'} style:--brightness={`${preferences.brightness}%`}>
 <header class="reader-bar">
  <a class="icon" href={`/titles/${manifest.titleSlug}`} aria-label="Тайтлға оралу">‹</a>
  <a class="reader-title" href={`/titles/${manifest.titleSlug}`}><strong>{manifest.titleName}</strong><span>{manifest.chapterLabel}-тарау · {manifest.team.name}</span></a>
  <div class="desktop-chapters">{#if manifest.previousId}<a class="icon" href={`/read/${manifest.previousId}`} aria-label="Алдыңғы тарау">‹</a>{/if}<button onclick={()=>openPanel('chapters')}>{manifest.chapterLabel}-тарау <span aria-hidden="true">⌄</span></button>{#if manifest.nextId}<a class="icon" href={`/read/${manifest.nextId}`} aria-label="Келесі тарау">›</a>{/if}</div>
  <button class="icon desktop-settings" onclick={()=>openPanel('settings')} aria-label="Оқу баптаулары">⚙</button>
 </header>
 {#if preview}<div class="preview-note">Жарияланбаған нұсқа · тек алдын ала қарау</div>{/if}
 {#if offlineTier}<div class="preview-note">Сақталған көшірме · желісіз оқу</div>{/if}
 {#if message}<div class="reader-message" role="status">{message}<button onclick={()=>message=''} aria-label="Хабарды жабу">×</button></div>{/if}
 <div class="reading-stage" bind:this={stage} style:max-width={`${preferences.fit==='height'&&preferences.mode==='pages'?Math.min(preferences.maxWidth,(viewport.height-20)*manifest.pages[pageIndex].width/manifest.pages[pageIndex].height):preferences.maxWidth}px`} style:aspect-ratio={`1/${sheet.total}`}>
  {#each painted as {item,state} (item.id)}
   <div class="fragment" style:top={`${item.top/sheet.total*100}%`} style:aspect-ratio={`${item.width}/${item.height}`}>
    {#if !mounted&&!offlineTier&&item===sheet.items[0]}<img bind:this={firstImage} src={item.url} width={item.width} height={item.height} alt={`${item.pageIndex+1}-бет`} fetchpriority="high" decoding="async"/>
    {:else if state?.status==='ready'}<img src={state.image?.url} width={item.width} height={item.height} alt={`${item.pageIndex+1}-бет`} decoding="async"/>
    {:else if state?.status==='error'}<div class="image-status" role="status"><p>{state.message}</p><button onclick={()=>queue?.retry(item.id)}>Қайта жүктеу</button></div>
    {:else}<div class="image-status"><span>{item.pageIndex+1}-бет</span><small>Жүктелуде…</small></div>{/if}
   </div>
  {/each}
  <button class="page-zone left-zone" aria-label={leftStep>0?'Келесі бет · сол жақ':'Алдыңғы бет · сол жақ'} disabled={!canStep(leftStep)} onclick={()=>step(leftStep)}></button>
  <button class="page-zone right-zone" aria-label={rightStep>0?'Келесі бет · оң жақ':'Алдыңғы бет · оң жақ'} disabled={!canStep(rightStep)} onclick={()=>step(rightStep)}></button>
 </div>
 {#if preferences.mode==='pages'&&pageIndex<manifest.pages.length-1}<section class="chapter-end"><button onclick={()=>step(1)}>Келесі бет →</button></section>{:else}<section class="chapter-end" aria-label="Тарауды аяқтау"><p>{manifest.chapterLabel}-тараудың соңы</p><span>{manifest.team.name}</span><div>{#if manifest.previousId}<a href={`/read/${manifest.previousId}`}>← Алдыңғы тарау</a>{/if}{#if manifest.nextId}<a class="next-chapter" href={`/read/${manifest.nextId}`}>Келесі тарау →</a>{:else}<a href={`/titles/${manifest.titleSlug}`}>Тайтлға оралу</a>{/if}</div></section>{#if community&&!preview&&!offlineTier}<div class="reader-discussion"><Discussion titleId={manifest.titleId} releaseId={manifest.releaseId} gateway={community}/></div>{/if}{/if}
 <nav class="reader-controls" aria-label="Оқу навигациясы"><button class="icon" onclick={()=>step(leftStep)} disabled={!canStep(leftStep)} aria-label={leftStep>0?'Келесі бет':'Алдыңғы бет'}>‹</button><button onclick={()=>openPanel('chapters')} aria-label="Тараулар тізімі"><span aria-hidden="true">☷</span><small>Тараулар</small></button><button onclick={()=>openPanel('settings')} aria-label="Оқу баптаулары"><span aria-hidden="true">⚙</span><small>Баптаулар</small></button><button class="icon" onclick={()=>step(rightStep)} disabled={!canStep(rightStep)} aria-label={rightStep>0?'Келесі бет':'Алдыңғы бет'}>›</button></nav>
 {#if preferences.numbers}<button class="page-counter" onclick={()=>openPanel('chapters')} aria-label="Ағымдағы бет">{current.pageIndex+1} / {manifest.pages.length}</button>{/if}
 <dialog bind:this={dialog} class="reader-drawer" onclose={()=>{if(panel)dismiss();}}>
  <header><h2>{panel==='settings'?'Оқу баптаулары':'Тараулар'}</h2><button class="icon" onclick={dismiss} aria-label="Панельді жабу">×</button></header>
  {#if panel==='settings'}
   <div class="drawer-body">
    {#if community&&!preview&&!offlineTier}<ReportForm gateway={community} titleId={manifest.titleId} releaseId={manifest.releaseId} pageId={manifest.pages[current.pageIndex].id} targetType="page" targetId={manifest.pages[current.pageIndex].id}/>{/if}
    {#if progressMessage}<p class="hint" role="status">{progressMessage}</p>{/if}{#if conflict}<div><p class="hint">Осы құрылғыдағы орын мен бұлттағы орын әртүрлі.</p><button onclick={()=>resolveProgress(true)}>Осы орынды сақтау</button><button onclick={()=>resolveProgress(false)}>Бұлттағы орынды қабылдау</button></div>{/if}
    <fieldset><legend>Оқу түрі</legend><div class="segments"><button class:selected={preferences.mode==='pages'} onclick={()=>mode('pages')}>Беттер</button><button class:selected={preferences.mode==='strip'} onclick={()=>mode('strip')}>Тұтас лента</button></div></fieldset>
    <label>Сурет сапасы<select value={preferences.quality} disabled={switching||!!offlineTier} onchange={event=>changeQuality(event.currentTarget.value as ReaderPreferences['quality'])}>{#each Object.entries(qualityNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
    <p class="hint">Автоматты: баяу желіде үнемді сапа. Жоғары сапа тек өз таңдауыңызбен жүктеледі.</p>
    {#if switching}<p role="status">Жаңа сапа жүктелуде. Ескі сурет сақталады…</p>{/if}
    <label>Тақырып<select bind:value={preferences.theme}><option value="dark">Қараңғы</option><option value="light">Ашық</option><option value="system">Жүйедегідей</option></select></label>
    <label>Суретті сыйдыру<select bind:value={preferences.fit}><option value="width">Ені бойынша</option><option value="height">Биіктігі бойынша</option></select></label>
    <label>Оқу бағыты<select bind:value={preferences.direction}><option value="auto">Тайтлға сәйкес</option><option value="ltr">Солдан оңға</option><option value="rtl">Оңнан солға</option></select></label>
    <p class="hint">Лентада: оң жақ — келесі бет, сол жақ — алдыңғы бет. Беттер режимінде бағыт жоғарыдағы таңдауға сәйкес.</p>
    <label class="range">Ең үлкен ені · {preferences.maxWidth}px<input type="range" min="320" max="1600" step="20" bind:value={preferences.maxWidth}/></label>
    <label class="range">Жарықтық · {preferences.brightness}%<input type="range" min="40" max="100" step="5" bind:value={preferences.brightness}/></label>
    <label class="check"><input type="checkbox" bind:checked={preferences.numbers}/>Бет нөмірін көрсету</label>
    <button onclick={()=>platform.fullscreen(shell).catch(()=>message='Толық экран қолжетімсіз')}>Толық экран</button>
    {#if !preview&&!offlineTier}<hr/>{#if downloads}<button onclick={saveOffline} disabled={downloadBusy||switching||!painted.some(entry=>entry.state?.status==='ready')}>Тарауды құрылғыға сақтау · {(manifest.pages.flatMap(source=>source.variants[tier]).reduce((sum,file)=>sum+file.bytes,0)/1024/1024).toFixed(1)} МБ</button>{#if downloadBusy}<button onclick={()=>downloadController?.abort()}>Жүктеуді тоқтату</button>{/if}{#if download}<p class="hint" role="status">{download.status==='ready'?'Толық сақталды':`${(download.bytes/1024/1024).toFixed(1)} / ${(download.total/1024/1024).toFixed(1)} МБ`} · <a href={`/offline-reader?release=${manifest.releaseId}`} data-sveltekit-reload>Офлайн ашу</a></p>{/if}<p class="hint">Көшірме осы браузерде ғана сақталады. Браузер оны өшіруі мүмкін. Сайттан алынған тарау бұрынғы офлайн көшірмеден автоматты өшпейді.</p>{/if}<button onclick={getOriginal}>Осы беттің түпнұсқасы</button>{#if original}<p class="hint">{(original.bytes/1024/1024).toFixed(1)} МБ. Жад пен интернет көбірек жұмсалады.</p><a class="original-link" href={original.url} target="_blank" rel="noreferrer">Түпнұсқаны бөлек ашу ↗</a>{/if}{/if}
   </div>
  {:else if panel==='chapters'}
   <div class="drawer-body"><label>Бетке өту<select value={current.pageIndex} onchange={event=>{void goPage(Number(event.currentTarget.value));dismiss();}}>{#each manifest.pages as source,index}<option value={index}>{index+1}-бет</option>{/each}</select></label>
    <p class="hint">Аударма тобы: {manifest.team.name}. Басқа топқа автоматты ауыспайды.</p>
    {#if !preview}<label>Тарауды іздеу<input type="search" bind:value={query} placeholder="Тарау нөмірі немесе атауы"/></label><button onclick={()=>reverse=!reverse}>{reverse?'Жаңасы алдымен ↓':'Ескісі алдымен ↑'}</button><div class="chapter-list">{#each shownChapters as chapter}{@const release=releaseFor(chapter)}{#if release}<a class:current={release.id===manifest.releaseId} href={`/read/${release.id}`}><strong>{chapter.label}-тарау</strong><span>{chapter.name}</span></a>{/if}{/each}</div>{#if chapterBusy}<p role="status">Жүктелуде…</p>{:else if cursor}<button onclick={loadChapters}>Тағы жүктеу</button>{/if}{/if}
   </div>
  {/if}
 </dialog>
</div>

<style>
 .reader{--bg:var(--reader-bg);--panel:var(--reader-panel);--text:var(--reader-text);--muted:var(--reader-muted);--line:var(--reader-line)}
 .reader-discussion{padding:0 20px}
 .reader{--reader-bg:#151515;--reader-panel:#202020;--reader-line:#363636;--reader-text:#ededed;--reader-muted:#b4b4b4;min-height:100dvh;background:var(--reader-bg);color:var(--reader-text);padding:52px 0 100px;line-height:1.4;color-scheme:dark}
 .reader.light{--reader-bg:#eee;--reader-panel:#fff;--reader-line:#ccc;--reader-text:#222;--reader-muted:#595959;color-scheme:light}
 @media(prefers-color-scheme:light){.reader.system{--reader-bg:#eee;--reader-panel:#fff;--reader-line:#ccc;--reader-text:#222;--reader-muted:#595959;color-scheme:light}}
 .reader button,.reader a{min-height:44px}.reader button{border:1px solid var(--reader-line);background:var(--reader-panel);color:var(--reader-text);border-radius:5px;padding:8px 12px}.reader .icon{display:grid;place-items:center;font-size:28px;width:44px;min-width:44px;padding:0;border:0;background:none}
 .reader-bar{height:52px;position:fixed;inset:0 0 auto;display:flex;align-items:center;gap:4px;padding:0 8px;background:var(--reader-panel);border-bottom:1px solid var(--reader-line);z-index:40}.reader-title{display:flex;flex-direction:column;justify-content:center;min-width:0;flex:1}.reader-title strong{font-size:14px;font-weight:550;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.reader-title span{font-size:11px;color:var(--reader-muted);white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.desktop-chapters,.desktop-settings{display:none!important}
 .reading-stage{position:relative;width:100%;margin:auto;min-height:200px;background:#181818;filter:brightness(var(--brightness))}.fragment{position:absolute;left:0;width:100%;overflow:hidden}.fragment img{display:block;width:100%;height:100%;object-fit:contain}.image-status{display:flex;flex-direction:column;align-items:center;justify-content:center;gap:10px;height:100%;min-height:150px;color:#bbb;background:#202020;font-size:14px;padding:20px}.image-status small{font-size:12px}.image-status button{border-color:#666}
 .reader .page-zone{position:absolute;top:0;bottom:0;width:30%;min-width:44px;border:0;border-radius:0;padding:0;background:transparent;touch-action:pan-y pinch-zoom;z-index:1}.left-zone{left:0}.right-zone{right:0}.page-zone:focus-visible{outline:2px solid #a2d560;outline-offset:-3px}.image-status button{position:relative;z-index:2}.reader-controls button:disabled{opacity:.35;cursor:default}
 .reader-controls{position:fixed;bottom:0;left:0;right:0;display:grid;grid-template-columns:44px 1fr 1fr 44px;align-items:center;padding:4px 12px calc(4px + env(safe-area-inset-bottom));border-top:1px solid var(--reader-line);background:var(--reader-panel);z-index:35}.reader-controls button{border:0;display:flex;flex-direction:column;align-items:center;gap:0;padding:0;line-height:1.1}.reader-controls button>span{font-size:24px;line-height:25px}.reader-controls small{font-size:10px}.page-counter{position:fixed;right:12px;bottom:72px;z-index:30;font-size:12px;font-variant-numeric:tabular-nums;min-width:64px;box-shadow:0 2px 6px #0003}
 .preview-note{text-align:center;padding:10px;background:#68551c;color:#fff;font-size:13px}.reader-message{position:sticky;top:52px;z-index:32;background:var(--reader-panel);border-bottom:1px solid var(--reader-line);display:flex;align-items:center;justify-content:center;padding:6px 12px;gap:10px;font-size:13px}.reader-message button{border:0;margin-left:auto}
 .chapter-end{padding:44px 16px;text-align:center}.chapter-end p{font-size:18px}.chapter-end>span{display:block;color:var(--reader-muted);font-size:13px;margin:8px 0 24px}.chapter-end>div{display:flex;flex-wrap:wrap;gap:12px;justify-content:center}.chapter-end a{display:flex;align-items:center;justify-content:center;border:1px solid var(--reader-line);padding:10px 18px;border-radius:5px}.chapter-end .next-chapter{background:#b4ee69;color:#17220d;border-color:#b4ee69}
 .reader-drawer{position:fixed;inset:0 0 0 auto;margin:0;width:100%;max-width:420px;height:100dvh;max-height:none;border:0;border-left:1px solid var(--reader-line);padding:0;background:var(--reader-panel);color:var(--reader-text);z-index:60}.reader-drawer::backdrop{background:#0008}.reader-drawer>header{position:sticky;top:0;background:var(--reader-panel);display:flex;align-items:center;justify-content:space-between;gap:8px;padding:4px 12px 4px 20px;border-bottom:1px solid var(--reader-line);z-index:1}.reader-drawer h2{font-size:17px;font-weight:600}.drawer-body{padding:20px;display:grid;gap:18px}.drawer-body label{display:grid;gap:8px;font-size:14px}.drawer-body select,.drawer-body input:not([type=checkbox]){background:var(--reader-bg);border-color:var(--reader-line);color:var(--reader-text);font-size:14px}.drawer-body fieldset{border:0;padding:0;margin:0;min-width:0}.drawer-body legend{font-size:14px;margin-bottom:8px}.segments{display:flex;gap:4px}.segments button{flex:1}.segments button.selected{background:#b4ee69;color:#17220d;border-color:#b4ee69}.hint{color:var(--reader-muted);font-size:12px;line-height:1.6}.range input{padding:0;accent-color:#79aa3d}.drawer-body .check{display:flex;align-items:center;min-height:44px;gap:12px}.check input{width:20px;height:20px;min-height:20px;accent-color:#79aa3d}.drawer-body hr{border:0;border-top:1px solid var(--reader-line);width:100%;margin:0}.original-link{display:flex;align-items:center;justify-content:center;border:1px solid var(--reader-line);font-size:14px}.chapter-list{display:grid}.chapter-list a{border-bottom:1px solid var(--reader-line);display:flex;flex-direction:column;justify-content:center;padding:10px 4px;gap:4px;font-size:14px}.chapter-list a span{font-size:12px;color:var(--reader-muted)}.chapter-list .current{color:#a2d560}
 @media(min-width:800px){.reader{padding-bottom:24px}.desktop-chapters{display:flex!important;align-items:center;gap:12px;position:absolute;left:50%;transform:translateX(-50%)}.desktop-chapters button{border:0}.reader-title{max-width:32%}.reader-bar .desktop-settings{display:grid!important;margin-left:auto}.reader-controls{left:auto;right:16px;bottom:calc(50% - 120px);width:48px;grid-template-columns:1fr;gap:6px;border:1px solid var(--reader-line);border-radius:6px;padding:4px 2px}.reader-controls small{display:none}.page-counter{bottom:calc(50% + 125px);right:12px}.reading-stage{width:calc(100% - 140px)}}
</style>
