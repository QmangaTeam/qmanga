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
 async function copyName(value:string){try{await navigator.clipboard.writeText(value);notice=text($locale,'Атауы көшірілді','Название скопировано');}catch{notice=text($locale,'Көшіру мүмкін болмады. Мәтінді белгіле.','Не удалось скопировать. Выдели текст вручную.');}}
</script>
<section class="metadata" aria-label={text($locale,'Тайтл туралы мәлімет','Сведения о тайтле')}>
 <dl class="summary-facts">
  <div><dt>{text($locale,'Түрі','Тип')}</dt><dd><a href={'/catalog?kind='+title.kind}>{names.kind[title.kind]}</a></dd></div>
  <div><dt>{text($locale,'Тайтл күйі','Статус тайтла')}</dt><dd><a href={'/catalog?original_status='+title.originalStatus}>{names.original[title.originalStatus]}</a></dd></div>
  <div><dt>{text($locale,'Аударма күйі','Статус перевода')}</dt><dd><a href={'/catalog?translation_status='+title.translationStatus}>{names.translation[title.translationStatus]}</a></dd></div>
  <div><dt>{text($locale,'Тараулар','Главы')}</dt><dd>{title.chapterCount??'—'}</dd></div>
  <div><dt>{text($locale,'Формат','Формат')}</dt><dd>{publication?.formats.length?publication.formats.join(', '):text($locale,'Белгісіз','Не указан')}</dd></div>
  {#if publication?.year}<div><dt>{text($locale,'Шыққан жылы','Год выпуска')}</dt><dd>{publication.year}</dd></div>{/if}
  {#if onStatistics}<div><dt>{text($locale,'Статистика','Статистика')}</dt><dd><button class="statistics" type="button" onclick={onStatistics}>{title.ratingCount} {text($locale,'баға','оценок')} <span aria-hidden="true">↗</span></button></dd></div>{/if}

  {#if publication?.publishers.length}<div><dt>{text($locale,'Баспалар','Издательства')}</dt><dd>{publication.publishers.join(', ')}</dd></div>{/if}
  {#if publication?.source}<div><dt>{text($locale,'Түпнегіз','Первоисточник')}</dt><dd>{publication.source}</dd></div>{/if}
  {#if publication?.country}<div><dt>{text($locale,'Шыққан елі','Страна производства')}</dt><dd>{publication.country}</dd></div>{/if}
  {#if title.authors?.length}<div><dt>{text($locale,'Авторлар','Авторы')}</dt><dd>{title.authors.join(', ')}</dd></div>{/if}
 </dl>
 {#if aliases.length}<h2>{text($locale,'Басқа атаулары','Другие названия тайтла')}</h2><ul class="aliases">{#each aliases as alias}<li><span>{alias}</span><button type="button" title={text($locale,'Атауын көшіру','Скопировать название')} aria-label={text($locale,'Атауын көшіру: ','Скопировать название: ')+alias} onclick={()=>copyName(alias)}><Icon name="copy" size={16}/></button></li>{/each}</ul>{/if}
 {#if publication?.links.length}<h2>{text($locale,'Сілтемелер','Ссылки')}</h2><ul class="links">{#each publication.links as link}{#if safeExternalURL(link.url)}<li><a href={link.url} target="_blank" rel="noopener noreferrer"><Icon name="link" size={16}/>{link.label}<span class="sr-only">{text($locale,' — жаңа бетте',' — в новой вкладке')}</span></a></li>{/if}{/each}</ul>{/if}
 {#if title.projects.length}<h2>{text($locale,'Аударма командалары','Команды перевода')}</h2><ul class="links">{#each title.projects as project}<li><a href={'/teams/'+encodeURIComponent(project.teamSlug)}><Icon name="users" size={16}/>{project.teamName}</a></li>{/each}</ul>{/if}
 {#if notice}<p role="status">{notice}</p>{/if}
</section>
<style>
 .metadata{padding:16px;border:1px solid var(--line);border-radius:10px;background:var(--panel);font-size:13px;min-width:0}
 dl,dd,ul{margin:0;padding:0}dl{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px 16px}dl>div{min-width:0}
 dt,h2{font-size:12px;font-weight:500;color:var(--muted);margin:0 0 4px;line-height:1.4}
 dd{font-weight:550;overflow-wrap:anywhere;line-height:1.5;min-height:20px}
 dd>a,.statistics{display:flex;align-items:flex-start;min-height:44px;color:var(--text);font:inherit;font-weight:550;text-decoration:none}
 .statistics{border:0;background:none;padding:0;gap:4px;cursor:pointer;text-align:left}
 h2{margin:20px 0 6px;font-size:13px}ul{list-style:none}.aliases li{display:flex;align-items:center;gap:8px;min-width:0}.aliases span{flex:1;overflow-wrap:anywhere}
 .aliases button{border:0;background:transparent;color:var(--muted);width:44px;min-height:44px;display:grid;place-items:center;flex:none}
 .links a{display:flex;align-items:center;gap:8px;min-height:44px;text-decoration:none;color:var(--text);overflow-wrap:anywhere}
 a:hover,.statistics:hover{text-decoration:underline}button:focus-visible,a:focus-visible{outline:2px solid var(--accent);outline-offset:3px}
 p{font-size:12px;margin-top:12px}
 @media(min-width:600px) and (max-width:959px){dl{grid-template-columns:repeat(3,minmax(0,1fr))}}
 @media(min-width:960px){dl{grid-template-columns:minmax(0,1fr);gap:16px}.metadata{padding:16px}}
</style>
