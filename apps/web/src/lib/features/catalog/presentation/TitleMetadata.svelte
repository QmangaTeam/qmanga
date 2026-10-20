<script lang="ts">
 import type {TitleDetail} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {safeExternalURL} from '../domain/publication';
 import {catalogNames} from './messages';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {title,onStatistics}:{title:TitleDetail;onStatistics?:()=>void}=$props();
 const {locale}=useLocale();
 const names=$derived(catalogNames($locale));
 const publication=$derived(title.publication);
 const aliases=$derived([...new Set([title.originalName,...title.aliases].filter(Boolean))]);
 let notice=$state('');
 let failedTeamAvatars=$state<string[]>([]);
 $effect(()=>{title.id;failedTeamAvatars=[];});
 async function copyName(value:string){try{await navigator.clipboard.writeText(value);notice=text($locale,'Атауы көшірілді','Название скопировано');}catch{notice=text($locale,'Көшіру мүмкін болмады. Мәтінді белгіле.','Не удалось скопировать. Выдели текст вручную.');}}
</script>
<section class="metadata" aria-label={text($locale,'Тайтл туралы мәлімет','Сведения о тайтле')}>
 <!-- svelte-ignore a11y_no_noninteractive_tabindex (Native horizontal scrolling supports keyboard users.) -->
 <div class="metadata-rail" tabindex="0" role="region" aria-label={text($locale,'Тайтл деректері','Факты о тайтле')}>
 <dl class="summary-facts">
  <div><dt><Icon name='catalog' size={13}/>{text($locale,'Түрі','Тип')}</dt><dd><a href={'/catalog?kind='+title.kind}>{names.kind[title.kind]}</a></dd></div>
  <div><dt><Icon name='file' size={13}/>{text($locale,'Тайтл күйі','Статус тайтла')}</dt><dd><a class="status-value" href={'/catalog?original_status='+title.originalStatus}><span class={'status-dot '+title.originalStatus} aria-hidden="true"></span>{names.original[title.originalStatus]}</a></dd></div>
  <div><dt><Icon name='users' size={13}/>{text($locale,'Аударма күйі','Статус перевода')}</dt><dd><a class="status-value" href={'/catalog?translation_status='+title.translationStatus}><span class={'status-dot '+title.translationStatus} aria-hidden="true"></span>{names.translation[title.translationStatus]}</a></dd></div>
  <div><dt><Icon name='catalog' size={13}/>{text($locale,'Формат','Формат')}</dt><dd>{publication?.formats.length?publication.formats.join(', '):text($locale,'Белгісіз','Не указан')}</dd></div>
  {#if publication?.year}<div><dt><Icon name='file' size={13}/>{text($locale,'Шыққан жылы','Год выпуска')}</dt><dd>{publication.year}</dd></div>{/if}
  {#if onStatistics}<div><dt><Icon name='chart' size={13}/>{text($locale,'Статистика','Статистика')}</dt><dd><button class="statistics" type="button" onclick={onStatistics}>{title.ratingCount} {text($locale,'баға','оценок')} <span aria-hidden="true">↗</span></button></dd></div>{/if}

  {#if publication?.source}<div><dt><Icon name='file' size={13}/>{text($locale,'Түпнегіз','Первоисточник')}</dt><dd>{publication.source}</dd></div>{/if}
  {#if publication?.country}<div><dt><Icon name='flag' size={13}/>{text($locale,'Шыққан елі','Страна производства')}</dt><dd>{publication.country}</dd></div>{/if}
 </dl>
 {#if aliases.length}<div class="metadata-group"><h2><Icon name="copy" size={13}/>{text($locale,'Басқа атаулары','Другие названия тайтла')}</h2><ul class="aliases">{#each aliases as alias}<li><span>{alias}</span><button type="button" title={text($locale,'Атауын көшіру','Скопировать название')} aria-label={text($locale,'Атауын көшіру: ','Скопировать название: ')+alias} onclick={()=>copyName(alias)}><Icon name="copy" size={16}/></button></li>{/each}</ul></div>{/if}
 {#if publication?.links.length}<div class="metadata-group"><h2><Icon name="link" size={13}/>{text($locale,'Сілтемелер','Ссылки')}</h2><ul class="links">{#each publication.links as link}{#if safeExternalURL(link.url)}<li><a href={link.url} target="_blank" rel="noopener noreferrer"><Icon name="link" size={16}/>{link.label}<span class="sr-only">{text($locale,' — жаңа бетте',' — в новой вкладке')}</span></a></li>{/if}{/each}</ul></div>{/if}
 </div>
</section>
 {#if title.authors?.length||publication?.publishers.length||title.projects.length}<ul class="contributors">
  {#each title.authors||[] as author}<li><span class="credit-chip"><Icon name="users" size={16}/><span><strong>{author}</strong><small>{text($locale,'Автор','Автор')}</small></span></span></li>{/each}
  {#each publication?.publishers||[] as publisher}<li><span class="credit-chip"><Icon name="file" size={16}/><span><strong>{publisher}</strong><small>{text($locale,'Баспа','Издатель')}</small></span></span></li>{/each}
  {#each title.projects as project}<li><a class="credit-chip" href={'/teams/'+encodeURIComponent(project.teamSlug)}><span class="team-avatar" aria-hidden="true">{#if project.teamId&&!failedTeamAvatars.includes(project.teamId)}<img src={'/media/avatars/team/'+encodeURIComponent(project.teamId)} alt="" width="32" height="32" loading="lazy" onerror={()=>failedTeamAvatars=[...failedTeamAvatars,project.teamId]}/>{:else}<Icon name="users" size={18}/>{/if}</span><span><strong>{project.teamName}</strong><small>{text($locale,'Аударма командасы','Команда перевода')}</small></span></a></li>{/each}
 </ul>{/if}
 {#if notice}<p role="status">{notice}</p>{/if}
<style>
 .metadata{padding:10px 12px;border:1px solid var(--line);border-radius:10px;background:var(--panel);font-size:12px;min-width:0}
 dl,dd,ul{margin:0;padding:0}dl{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px 16px}dl>div{min-width:0}
 dt,h2{font-size:12px;font-weight:500;color:var(--muted);margin:0 0 4px;line-height:1.4}
 dd{font-weight:550;overflow-wrap:anywhere;line-height:1.5;min-height:20px}
 dd>a,.statistics{display:flex;align-items:flex-start;min-height:44px;color:var(--text);font:inherit;font-weight:550;text-decoration:none}
 .statistics{border:0;background:none;padding:0;gap:4px;cursor:pointer;text-align:left}
 h2{margin:20px 0 6px;font-size:13px}ul{list-style:none}.aliases li{display:flex;align-items:center;gap:8px;min-width:0}.aliases span{flex:1;overflow-wrap:anywhere}
 .aliases button{border:0;background:transparent;color:var(--muted);width:44px;min-height:44px;display:grid;place-items:center;flex:none}
 .links a{position:relative;display:flex;align-items:center;gap:8px;min-height:44px;text-decoration:none;color:var(--text);overflow-wrap:anywhere}
 a:hover,.statistics:hover{text-decoration:underline}button:focus-visible,a:focus-visible{outline:2px solid var(--accent);outline-offset:3px}
 p{font-size:12px;margin-top:12px}
 dt,h2{display:flex;align-items:center;gap:5px}
 .status-value{gap:5px}.status-dot{width:6px;height:6px;margin-top:6px;border-radius:50%;flex:none;background:var(--muted)}
 .status-dot.ongoing{background:#3ac66d}.status-dot.completed{background:#4894ee}.status-dot.hiatus,.status-dot.paused{background:#e3a52d}.status-dot.cancelled,.status-dot.stopped{background:#e55b60}
 .team-avatar{width:32px;height:32px;flex:none;display:grid;place-items:center;overflow:hidden;border-radius:5px;background:var(--bg)}.team-avatar img{width:100%;height:100%;object-fit:cover}
 @media(max-width:959px){
  .metadata-rail{display:flex;gap:14px;overflow-x:auto;scroll-snap-type:x proximity;scroll-padding-inline:0;overscroll-behavior-x:contain;scrollbar-width:none;padding-bottom:4px}
  dl{display:contents}dl>div,.metadata-group{flex:0 0 auto;width:max-content;min-width:88px;max-width:156px;scroll-snap-align:start}
  .metadata-group:has(.aliases){max-width:none}.aliases{display:flex;gap:12px}.aliases li{width:230px;flex:0 0 230px}
  h2{margin:0 0 4px}.links a{align-items:flex-start;padding-block:4px}.aliases li{align-items:flex-start}
 }
 .metadata-rail::-webkit-scrollbar{display:none}
 .metadata-rail:focus-visible{outline:2px solid var(--accent);outline-offset:3px}
 .contributors{display:flex;flex-wrap:wrap;gap:8px;margin-top:10px}
 .contributors li{min-width:0;max-width:100%}.credit-chip{display:flex;align-items:center;gap:8px;min-height:44px;max-width:100%;padding:6px 8px;border:1px solid var(--line);border-radius:6px;background:var(--panel);color:var(--text);text-decoration:none}
 .credit-chip>span{min-width:0}.credit-chip strong{display:block;font-size:12px;font-weight:500;overflow-wrap:anywhere}.credit-chip small{display:block;color:var(--muted);font-size:11px;line-height:1.4}
 @media(min-width:960px){dl{grid-template-columns:minmax(0,1fr);gap:16px}.metadata{padding:16px}}
</style>
