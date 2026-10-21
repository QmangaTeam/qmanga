<script lang="ts">
 import type {ReleaseUpdate} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 let {items,compact=false}:{items:ReleaseUpdate[];compact?:boolean}=$props();
 const {locale}=useLocale();
</script>
<div class="updates" class:compact style:--update-columns={Math.min(3,Math.max(1,items.length))} style:--update-tablet-columns={Math.min(2,Math.max(1,items.length))}>
 {#each items as item (item.releaseId)}
  <article>
   <a class="thumbnail" href={'/titles/'+encodeURIComponent(item.titleSlug)} aria-label={item.titleName}>{#if item.coverURL}<img src={item.coverURL} alt="" width="112" height="160" loading="lazy" decoding="async"/>{:else}<Icon name="catalog"/>{/if}</a>
   <div class="update-content">
    <a class="title" href={'/titles/'+encodeURIComponent(item.titleSlug)}><span>{item.titleName}{item.demo?text($locale,' · Үлгі',' · Пример'):''}</span></a>
    {#if !compact}<p class="team">{item.teamName}</p>{/if}
    <div class="release-line"><a class="chapter" href={'/read/'+encodeURIComponent(item.releaseId)}>{text($locale,item.chapterLabel+'-тарау','Глава '+item.chapterLabel)}<Icon name="right" size={14}/></a><time datetime={item.publishedAt}>{new Date(item.publishedAt).toLocaleDateString($locale==='ru'?'ru-RU':'kk-KZ')}</time></div>
   </div>
  </article>
 {:else}<StateMessage kind="empty" title={text($locale,'Әзірге жаңа тарау жоқ','Новых глав пока нет')}/>{/each}
</div>
<style>
 .updates{display:grid;gap:12px 24px;margin:0 0 24px}article{display:flex;align-items:center;gap:12px;padding:12px;border:1px solid var(--line);border-radius:12px;background:var(--panel);min-width:0}.thumbnail{display:grid;place-items:center;flex:0 0 64px;height:96px;background:var(--panel);border-radius:6px;overflow:hidden;color:var(--muted)}img{width:100%;height:100%;object-fit:cover}.update-content{flex:1;min-width:0}.title{min-height:44px;display:flex;align-items:center;font-size:14px;font-weight:600;line-height:1.35;overflow-wrap:anywhere}.title span{display:-webkit-box;-webkit-line-clamp:2;line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}.team{font-size:12px;color:var(--muted);white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.release-line{display:flex;align-items:center;justify-content:space-between;gap:8px}.chapter{border-radius:7px;background:var(--panel);border:1px solid var(--line);padding:0 10px;display:flex;align-items:center;gap:4px;min-height:44px;min-width:44px;font-size:13px;color:var(--text);overflow-wrap:anywhere}time{font-size:11px;color:var(--muted);white-space:nowrap}.title:hover,.chapter:hover{color:var(--accent)}@media(min-width:640px){.updates{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(min-width:1100px){.updates{grid-template-columns:repeat(3,minmax(0,1fr))}}
.updates.compact{margin-bottom:0}@media(min-width:640px){.updates.compact{grid-template-columns:repeat(var(--update-tablet-columns),minmax(0,1fr))}}@media(min-width:1100px){.updates.compact{grid-template-columns:repeat(var(--update-columns),minmax(0,1fr))}}
</style>
