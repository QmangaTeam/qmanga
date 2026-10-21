<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 const {locale}=useLocale();
 let {title,releases,readLabel}:{title:string;releases:{id:string;teamName:string;teamSlug:string;publishedAt?:string}[];readLabel:string}=$props();
 let query=$state('');
 const filtered=$derived(releases.filter(release=>release.teamName.toLocaleLowerCase($locale).includes(query.trim().toLocaleLowerCase($locale))));
 const date=(value:string)=>new Date(value).toLocaleDateString($locale==='ru'?'ru-RU':'kk-KZ',{timeZone:'UTC'});
</script>
<article class="chapter-row">
 {#if releases.length===1}{@const release=releases[0]}
  <div class="single-release"><h3><a class="chapter-link" href={'/read/'+encodeURIComponent(release.id)} title={title}>{title}<Icon name="right" size={16}/></a></h3><a class="team" href={'/teams/'+encodeURIComponent(release.teamSlug)}>{release.teamName}</a>{#if release.publishedAt}<time datetime={release.publishedAt}>{date(release.publishedAt)}</time>{/if}</div>
 {:else if releases.length>1}
  <details class="edition-picker"><summary><h3>{title}</h3><span class="version-count">{releases.length} {text($locale,'нұсқа','версий')}</span><Icon name="down" size={16}/></summary>
   <div class="edition-panel"><p>{text($locale,'Аударманы таңдаңыз','Выберите перевод')}</p>
    {#if releases.length>5}<label class="edition-search"><Icon name="search" size={16}/><input type="search" bind:value={query} placeholder={text($locale,'Команданы іздеу','Найти команду')} aria-label={text($locale,'Аударма командасын іздеу','Поиск команды перевода')}/></label>{/if}
    <div class="edition-list" role="list" aria-label={text($locale,'Тарау аудармалары','Переводы главы')}>
     {#each filtered as release (release.id)}<div class="edition" role="listitem"><a class="edition-read" href={'/read/'+encodeURIComponent(release.id)}><span class="edition-name"><strong>{release.teamName}</strong>{#if release.publishedAt}<time datetime={release.publishedAt}>{date(release.publishedAt)}</time>{/if}</span><span class="read-caption">{readLabel}<Icon name="right" size={14}/></span></a><a class="team-profile" href={'/teams/'+encodeURIComponent(release.teamSlug)} aria-label={text($locale,'Команда: ','Команда: ')+release.teamName} title={release.teamName}><Icon name="users" size={18}/></a></div>
     {:else}<p class="empty-editions">{text($locale,'Команда табылмады','Команда не найдена')}</p>{/each}
    </div>
   </div>
  </details>
 {:else}<h3 class="unavailable">{title}</h3>{/if}
</article>
<style>
 .chapter-row{min-width:0;border:1px solid var(--line);border-radius:8px;background:var(--panel);overflow:hidden}h3{font-size:14px;font-weight:550;line-height:1.4;margin:0;min-width:0;overflow-wrap:anywhere}
 .chapter-link{display:flex;align-items:center;justify-content:space-between;gap:8px;min-height:44px;padding:6px 10px;color:var(--text);text-decoration:none}.chapter-link :global(svg){flex:none;color:var(--muted)}
 .single-release{display:grid;grid-template-columns:minmax(0,1fr) auto;align-items:center}.single-release h3{grid-column:1/-1}.team{display:flex;align-items:center;min-height:44px;min-width:0;padding:4px 10px;font-size:12px;color:var(--muted);overflow-wrap:anywhere;text-decoration:none}
 time{font-size:11px;color:var(--muted);white-space:nowrap}.single-release>time{padding:4px 10px}
 summary{display:flex;align-items:center;gap:10px;min-height:48px;padding:8px 10px;cursor:pointer;list-style:none}summary::-webkit-details-marker{display:none}summary h3{flex:1;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.edition-picker[open] summary h3{white-space:normal}.version-count{font-size:11px;color:var(--muted);white-space:nowrap;background:var(--raised);border-radius:4px;padding:3px 6px}summary :global(svg){flex:none;color:var(--muted)}.edition-picker[open] summary :global(svg){transform:rotate(180deg)}
 .edition-panel{border-top:1px solid var(--line);padding:8px 10px 10px}.edition-panel>p{font-size:12px;color:var(--muted);margin:0 0 8px}.edition-search{display:flex;align-items:center;gap:6px;border:1px solid var(--line);border-radius:6px;padding-inline:8px;margin-bottom:8px;color:var(--muted)}.edition-search input{width:100%;min-width:0;min-height:44px;border:0;background:transparent;padding-inline:0}
 .edition-list{max-height:min(320px,40dvh);overflow:auto;overscroll-behavior:contain;scrollbar-width:thin}.edition{display:flex;align-items:stretch;gap:4px;border-bottom:1px solid var(--line)}.edition:last-child{border-bottom:0}.edition-read{display:flex;align-items:center;gap:10px;flex:1;min-width:0;min-height:48px;padding:6px 4px;color:var(--text);text-decoration:none}.edition-name{min-width:0;flex:1}.edition-name strong{display:block;font-size:13px;font-weight:500;overflow-wrap:anywhere}.edition-name time{display:block;margin-top:2px}.read-caption{display:flex;align-items:center;gap:4px;flex:none;font-size:12px;color:var(--accent)}.team-profile{display:grid;place-items:center;width:44px;min-height:44px;flex:none;color:var(--muted)}.empty-editions,.unavailable{padding:10px;font-size:13px;color:var(--muted)}a:hover{color:var(--accent)}summary:focus-visible,a:focus-visible{outline:2px solid var(--accent);outline-offset:-2px}
 @media(min-width:700px){.single-release{grid-template-columns:minmax(0,1fr) minmax(110px,.45fr) auto}.single-release h3{grid-column:auto}.chapter-link{min-width:0}.chapter-link{display:block;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;line-height:32px}.chapter-link :global(svg){display:none}}
</style>
