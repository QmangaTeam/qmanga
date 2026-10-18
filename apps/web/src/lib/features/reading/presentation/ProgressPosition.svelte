<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import type {Progress} from '../application/progress';
 let {progress,chapterLabel='',pagePosition=0,updatedAt=0}:{progress:Progress;chapterLabel?:string;pagePosition?:number;updatedAt?:number}=$props();
 const {locale}=useLocale();
 const chapter=$derived(chapterLabel||progress.chapterLabel);
 const position=$derived(pagePosition||progress.pagePosition);
 const timestamp=$derived(updatedAt||progress.updatedAt||0);
</script>
<span class="position">{chapter?text($locale,chapter+'-тарау','Глава '+chapter):text($locale,'Тарау белгісіз','Глава неизвестна')} · {position?text($locale,position+'-бет','стр. '+position):text($locale,'Бет белгісіз','Страница неизвестна')}{#if progress.teamName} · {progress.teamName}{/if}{#if timestamp>0} · <time datetime={new Date(timestamp).toISOString()}>{new Date(timestamp).toLocaleString($locale==='ru'?'ru-RU':'kk-KZ',{timeZone:'UTC',year:'numeric',month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit'})} UTC</time>{/if}</span>
<style>.position{display:block;color:var(--muted);font-size:13px;line-height:1.6;overflow-wrap:anywhere}</style>
