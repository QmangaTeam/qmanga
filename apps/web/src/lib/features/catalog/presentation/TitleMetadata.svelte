<script lang="ts">
 import type {TitleDetail} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {safeExternalURL} from '../domain/publication';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {title}:{title:TitleDetail}=$props();
 const {locale}=useLocale();
 const publication=$derived(title.publication);
 const aliases=$derived([...new Set([title.originalName,...title.aliases].filter(Boolean))]);
 let notice=$state('');
 async function copyName(value:string){try{await navigator.clipboard.writeText(value);notice=text($locale,'Атауы көшірілді','Название скопировано');}catch{notice=text($locale,'Көшіру мүмкін болмады. Мәтінді белгіле.','Не удалось скопировать. Выдели текст вручную.');}}
</script>
<section class="metadata" aria-label={text($locale,'Тайтл туралы мәлімет','Сведения о тайтле')}>
 <dl>
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
 .metadata{padding:16px;border:1px solid var(--line);border-radius:12px;background:var(--panel);font-size:13px;min-width:0}.metadata:empty{display:none}dl,dd,ul{margin:0;padding:0}dl>div{margin-bottom:16px}dt,h2{font-size:13px;font-weight:500;color:var(--muted);margin:0 0 6px}h2:not(:first-child){margin-top:18px}dd{font-weight:550;overflow-wrap:anywhere}ul{list-style:none}.aliases li{display:flex;align-items:center;gap:8px;min-width:0}.aliases span{flex:1;overflow-wrap:anywhere}.aliases button{border:0;background:transparent;color:var(--muted);width:44px;min-height:44px;display:grid;place-items:center;flex:none}.links a{display:flex;align-items:center;gap:8px;min-height:44px;text-decoration:none;color:var(--text);overflow-wrap:anywhere}p{font-size:12px;margin-top:12px}
</style>
