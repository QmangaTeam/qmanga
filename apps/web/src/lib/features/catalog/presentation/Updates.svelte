<script lang="ts">
 import type {ReleaseUpdate} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 let {items}:{items:ReleaseUpdate[]}=$props();
 const {locale}=useLocale();
</script>
<div class="updates">{#each items as item (item.releaseId)}<article><div><a class="title" href={'/titles/'+encodeURIComponent(item.titleSlug)}>{item.titleName}{item.demo?text($locale,' · Үлгі',' · Пример'):''}</a><a class="team" href={'/teams/'+encodeURIComponent(item.teamSlug)}>{item.teamName}</a><time datetime={item.publishedAt}>{new Date(item.publishedAt).toLocaleDateString($locale==='ru'?'ru-RU':'kk-KZ')}</time></div><a class="button" href={'/read/'+encodeURIComponent(item.releaseId)}>{text($locale,item.chapterLabel+'-тарау','Глава '+item.chapterLabel)}</a></article>{:else}<StateMessage kind="empty" title={text($locale,'Әзірге жаңа тарау жоқ','Новых глав пока нет')}/>{/each}</div>
<style>.updates{margin:16px 0 32px}article{display:flex;align-items:center;gap:12px;border-bottom:1px solid var(--line);padding:12px 0}article>div{flex:1;min-width:0}a{min-height:44px;display:flex;align-items:center;overflow-wrap:anywhere}.title{font-weight:600}.team{font-size:14px;color:var(--muted)}time{font-size:12px;color:var(--muted)}.button{flex-shrink:0;max-width:45%}</style>
