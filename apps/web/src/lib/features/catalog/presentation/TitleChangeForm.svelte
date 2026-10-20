<script lang="ts">
 import {fieldValidation} from '$lib/shared/presentation/form-validation';
 import {onMount} from 'svelte';
 import {beforeNavigate} from '$app/navigation';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {noticeText,failureCode} from '$lib/shared/presentation/copy';
 import type {TitleChangesGateway,TitleChangeContext,TitleChangeProposal,TitleChange,ChangeFiles} from '../application/title-changes';
 import type {Taxonomy} from '../application/catalog';
 import PublicationFields from './PublicationFields.svelte';
 import TaxonomySelection from './TaxonomySelection.svelte';
 import {catalogNames} from './messages';
 let {gateway,projectId,coverPageId=''}:{gateway:TitleChangesGateway;projectId:string;coverPageId?:string}=$props();
 const {locale}=useLocale();const names=$derived(catalogNames($locale));
 const taxonomyFields=['genre_ids','category_ids'] as const;
 let actor=$state(''),context=$state<TitleChangeContext|null>(null),proposal=$state<TitleChangeProposal|null>(null),taxonomy=$state<Taxonomy>({genres:[],categories:[],collections:[]}),history=$state<TitleChange[]>([]),hasMore=$state(false),loading=$state(true),busy=$state(false),error=$state(''),saved=$state(false),dirty=$state(false),links=$state(''),files=$state<ChangeFiles>({evidence:[]});
 let generation=0;
 const status=$derived({pending:text($locale,'Қаралуда','На рассмотрении'),approved:text($locale,'Мақұлданды','Одобрено'),rejected:text($locale,'Қайтарылды','Возвращено')});
 async function load(){const run=++generation;loading=true;actor='';context=null;proposal=null;history=[];files={evidence:[]};links='';dirty=false;error='';try{const account=await gateway.currentAccount();if(!account)throw Error('unauthorized');const [current,taxa,requests]=await Promise.all([gateway.context(account,projectId),gateway.taxonomy(),gateway.list(account,projectId)]);if(run!==generation)return;actor=account;context=current;taxonomy=taxa;history=requests.items;hasMore=requests.has_more;proposal={base_version:current.base_version,data:structuredClone(current.snapshot.data),reason:'',links:[],cover_action:coverPageId?'replace':'keep',banner_action:'keep',cover_page_id:coverPageId};}catch(e){if(run===generation)error=failureCode(e);}finally{if(run===generation)loading=false;}}
 async function refreshDraft(){if(dirty&&!window.confirm(text($locale,'Деректерді жаңартып, сақталмаған өзгерістерді жою керек пе?','Обновить данные и удалить несохранённые изменения?')))return;saved=false;await load();}
 async function more(){if(busy)return;busy=true;const run=generation;try{const result=await gateway.list(actor,projectId,history.at(-1)?.id);if(run===generation){history=[...history,...result.items];hasMore=result.has_more;}}catch(e){if(run===generation)error=failureCode(e);}finally{if(run===generation)busy=false;}}
 async function submit(event:SubmitEvent){event.preventDefault();if(!proposal||busy)return;busy=true;error='';const run=generation;try{if(proposal.cover_action!=='replace'){proposal.cover_page_id='';files.cover=undefined;}if(proposal.banner_action!=='replace')files.banner=undefined;proposal.links=links.split('\n').map(x=>x.trim()).filter(Boolean);await gateway.submit(actor,projectId,$state.snapshot(proposal),$state.snapshot(files));if(run!==generation)return;dirty=false;saved=true;busy=false;await load();}catch(e){if(run===generation)error=failureCode(e);}finally{if(run===generation)busy=false;}}
 function artwork(event:Event,kind:'cover'|'banner'){const file=(event.currentTarget as HTMLInputElement).files?.[0];files={...files,[kind]:file};if(proposal){if(kind==='cover'){proposal.cover_page_id='';proposal.cover_action=file?'replace':'keep';}else proposal.banner_action=file?'replace':'keep';}dirty=true;}
 beforeNavigate(({cancel})=>{if(dirty&&!window.confirm(text($locale,'Сақталмаған өзгерістерді қалдыру керек пе?','Уйти без сохранения изменений?')))cancel();});
 onMount(()=>{void load();const stop=gateway.watchAccount(()=>{busy=false;saved=false;links='';void load();});const unload=(e:BeforeUnloadEvent)=>{if(dirty){e.preventDefault();e.returnValue='';}};window.addEventListener('beforeunload',unload);return()=>{generation++;stop();window.removeEventListener('beforeunload',unload);};});
</script>
<svelte:head><title>{text($locale,'Жоба деректерін өзгерту','Изменить данные проекта')} — Qmanga</title></svelte:head>
<a class="back" href={'/studio/projects/'+encodeURIComponent(projectId)}>{text($locale,'Жобаға оралу','Вернуться к проекту')}</a>
<h1>{text($locale,'Жоба деректерін өзгерту','Изменить данные проекта')}</h1>
<p class="muted">{text($locale,'Өзгерістер модерациядан кейін барлық командалар үшін тайтл карточкасына қолданылады.','После модерации изменения применятся к общей карточке тайтла для всех команд.')}</p>
{#if loading}<p role="status">{text($locale,'Жүктелуде…','Загрузка…')}</p>{/if}
{#if error}<p class="notice" role="alert">{noticeText($locale,error)}</p><button class="button subtle" disabled={busy||loading} onclick={refreshDraft}>{text($locale,'Қайта жүктеу','Обновить данные')}</button>{/if}
{#if saved}<p class="notice" role="status">{text($locale,'Өтінім модерацияға жіберілді.','Заявка отправлена на модерацию.')}</p>{/if}
{#if proposal&&context}
 <h2>{context.title_name}</h2>
 {#if history.some(x=>x.status==='pending')}<p class="notice">{text($locale,'Осы жобаның өтінімі қаралуда. Шешімді күтіңіз.','Заявка этого проекта уже рассматривается. Дождитесь решения.')}</p>{:else}
 <form use:fieldValidation={{locale:$locale}} onsubmit={submit} oninput={()=>dirty=true}>
 <fieldset disabled={busy}>
 <section class="panel fields"><h2>{text($locale,'Тайтл туралы','Сведения о тайтле')}</h2>
 <label>{text($locale,'Атауы','Название')}<input bind:value={proposal.data.name} maxlength="200" required/></label>
 <label>{text($locale,'Түпнұсқа атауы','Оригинальное название')}<input bind:value={proposal.data.original_name} maxlength="200"/></label>
 <label>{text($locale,'Басқа атаулар, әр жолға біреуі','Другие названия, по одному на строке')}<textarea rows="3" value={proposal.data.aliases.join('\n')} onchange={e=>proposal!.data.aliases=e.currentTarget.value.split('\n').map(x=>x.trim()).filter(Boolean)}></textarea></label>
 <label>{text($locale,'Авторлар, әр жолға біреуі','Авторы, по одному на строке')}<textarea rows="2" value={proposal.data.authors.join('\n')} onchange={e=>proposal!.data.authors=e.currentTarget.value.split('\n').map(x=>x.trim()).filter(Boolean)}></textarea></label>
 <div class="columns"><label>{text($locale,'Түрі','Тип')}<select bind:value={proposal.data.kind}>{#each Object.entries(names.kind) as [value,label]}<option {value}>{label}</option>{/each}</select></label>
 <label>{text($locale,'Тайтл күйі','Статус тайтла')}<select bind:value={proposal.data.original_status}>{#each Object.entries(names.original) as [value,label]}<option {value}>{label}</option>{/each}</select></label>
 <label>{text($locale,'Аударма күйі','Статус перевода')}<select bind:value={proposal.data.translation_status}>{#each Object.entries(names.translation) as [value,label]}<option {value}>{label}</option>{/each}</select></label>
 <label>{text($locale,'Оқу бағыты','Направление чтения')}<select bind:value={proposal.data.direction}><option value="ltr">{text($locale,'Солдан оңға','Слева направо')}</option><option value="rtl">{text($locale,'Оңнан солға','Справа налево')}</option></select></label>
 <label>{text($locale,'Жас шектеуі','Возрастная маркировка')}<select bind:value={proposal.data.age_rating}><option value={0}>{text($locale,'Көрсетілмеген','Не указана')}</option>{#each [12,16,18] as age}<option value={age}>{age}+</option>{/each}</select></label></div>
 <label>{text($locale,'Сипаттама','Описание')}<textarea rows="6" maxlength="10000" bind:value={proposal.data.description}></textarea></label>
 <div class="columns">{#each taxonomyFields as kind}<TaxonomySelection label={kind==='genre_ids'?text($locale,'Жанрлар','Жанры'):text($locale,'Санаттар','Категории')} items={kind==='genre_ids'?taxonomy.genres:taxonomy.categories} selected={proposal.data[kind]} onchange={ids=>{proposal!.data[kind]=ids;dirty=true;}}/>{/each}</div>
 <PublicationFields bind:value={proposal.data.publication}/>
 </section>
 <section class="panel fields"><h2>{text($locale,'Мұқаба және баннер','Обложка и баннер')}</h2><p class="muted">{text($locale,'Суреттер де модерациядан өтеді. Баннер болмаса, мұқаба көрсетіледі.','Изображения тоже проходят модерацию. Если баннера нет, используется обложка.')}</p>
 {#each ['cover','banner'] as kind}<div class="artwork"><h3>{kind==='cover'?text($locale,'Мұқаба','Обложка'):text($locale,'Баннер','Баннер')}</h3>
 {#if context.snapshot[kind==='cover'?'cover_id':'banner_id']}<img src={'/api/v1/media/'+encodeURIComponent(context.snapshot[kind==='cover'?'cover_id':'banner_id'])} alt={kind==='cover'?text($locale,'Қазіргі мұқаба','Текущая обложка'):text($locale,'Қазіргі баннер','Текущий баннер')} class:banner={kind==='banner'}/>{/if}
 <label>{text($locale,'Әрекет','Действие')}<select bind:value={proposal[kind==='cover'?'cover_action':'banner_action']}><option value="keep">{text($locale,'Өзгеріссіз','Без изменений')}</option><option value="replace">{text($locale,'Ауыстыру','Заменить')}</option><option value="remove" disabled={!context.snapshot[kind==='cover'?'cover_id':'banner_id']}>{text($locale,'Алып тастау','Удалить')}</option></select></label>
 {#if proposal[kind==='cover'?'cover_action':'banner_action']==='replace'}
 {#if kind==='cover'&&proposal.cover_page_id}<p>{text($locale,'Таңдалған тарау беті мұқаба ретінде ұсынылады.','Выбранная страница главы будет предложена как обложка.')}</p>{/if}
 <label>{text($locale,'JPEG, PNG, WebP · 25 MiB-қа дейін','JPEG, PNG, WebP · до 25 MiB')}<input type="file" accept="image/jpeg,image/png,image/webp" required={kind==='banner'||!proposal.cover_page_id} onchange={e=>artwork(e,kind as 'cover'|'banner')}/></label>{/if}</div>{/each}
 </section>
 <section class="panel fields"><h2>{text($locale,'Негіздеме және дәлелдер','Обоснование и доказательства')}</h2>
 <label>{text($locale,'Нені және не үшін өзгертесіз?','Что и почему нужно изменить?')}<textarea bind:value={proposal.reason} minlength="3" maxlength="2000" rows="4" required></textarea></label>
 <label>{text($locale,'Дереккөз сілтемелері, әр жолға біреуі','Ссылки на источники, по одной на строке')}<textarea bind:value={links} rows="3" maxlength="20000" required={files.evidence.length===0}></textarea></label>
 <label>{text($locale,'Дәлел суреттері · 5 файлға дейін, әрқайсысы 5 MiB','Изображения доказательств · до 5 файлов, каждое до 5 MiB')}<input type="file" multiple accept="image/jpeg,image/png,image/webp" onchange={e=>files.evidence=Array.from(e.currentTarget.files||[])}/></label>
 <p class="muted">{text($locale,'Кемінде бір сілтеме немесе дәлел суреті қажет. Ұсынылған мұқаба не баннер өздігінен дәлел болып саналмайды.','Нужна хотя бы одна ссылка или изображение доказательства. Предлагаемая обложка или баннер сами по себе доказательством не считаются.')}</p>
 </section>
 <button class="button primary" type="submit" disabled={busy}>{busy?text($locale,'Жіберілуде…','Отправка…'):text($locale,'Модерацияға жіберу','Отправить на модерацию')}</button>
 </fieldset></form>{/if}
 <section class="history"><h2>{text($locale,'Өтінімдер тарихы','История заявок')}</h2>{#each history as request}<article class="panel"><a href={'/studio/title-changes/'+encodeURIComponent(request.id)}>{request.title_name}</a><p>{status[request.status]}</p><p>{request.reason}</p>{#if request.resolution}<p>{request.resolution}</p>{/if}</article>{:else}<p class="muted">{text($locale,'Әзірге өтінім жоқ.','Заявок пока нет.')}</p>{/each}{#if hasMore}<button class="button subtle" disabled={busy} onclick={more}>{text($locale,'Тағы жүктеу','Загрузить ещё')}</button>{/if}</section>
{/if}
<style>h1{font-size:24px}.fields{display:grid;gap:16px;margin-block:16px}.fields h2,.fields h3{margin:0}fieldset{border:0;padding:0;min-width:0}label{display:grid;gap:8px;font-size:14px;min-width:0}input,textarea,select{max-width:100%;min-width:0}.columns{display:grid;gap:16px}.artwork{display:grid;gap:12px}.artwork img{width:100px;aspect-ratio:2/3;object-fit:cover;border-radius:6px}.artwork img.banner{width:min(100%,400px);aspect-ratio:3/1}.history{margin-top:32px}.history article{margin-block:12px;overflow-wrap:anywhere}.history p{white-space:pre-wrap}.notice{padding:12px;border:1px solid var(--line);border-radius:8px}.muted{line-height:1.5}@media(min-width:700px){.columns{grid-template-columns:repeat(2,minmax(0,1fr))}}</style>
