<script lang="ts">
 import {onMount} from 'svelte';
 import type {Summary} from '../application/catalog';
 import {kindNames} from '../domain/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import IconButton from '$lib/shared/presentation/ui/IconButton.svelte';
 import CoverCard from './CoverCard.svelte';
 let {items,label,eager=false}:{items:Summary[];label:string;eager?:boolean}=$props();
 const {locale}=useLocale();
 let shelf:HTMLDivElement;
 let canGoBack=$state(false),canGoForward=$state(false);
 function measure(){canGoBack=shelf.scrollLeft>1;canGoForward=shelf.scrollLeft+shelf.clientWidth<shelf.scrollWidth-1;}
 function move(direction:number){shelf.scrollBy({left:direction*shelf.clientWidth*.8,behavior:'instant'});}
 onMount(()=>{const observer=new ResizeObserver(measure);observer.observe(shelf);measure();return()=>observer.disconnect();});
</script>
<div class="shelf-wrap">
 <div class="cover-shelf" role="region" aria-label={label} bind:this={shelf} onscroll={measure}>
  {#each items as title,index (title.id)}
   <div class="shelf-item"><CoverCard href={'/titles/'+encodeURIComponent(title.slug)} name={title.name} coverURL={title.coverURL} metadata={kindNames[title.kind]} tag={title.demo?text($locale,'Үлгі','Пример'):''} score={title.score} ratingCount={title.ratingCount} fallback={text($locale,'Мұқаба жоқ','Нет обложки')} priority={eager&&index<3}/></div>
  {/each}
 </div>
 {#if canGoBack||canGoForward}<div class="shelf-controls"><IconButton name="left" label={text($locale,'Алдыңғы тайтлдар','Предыдущие тайтлы')} disabled={!canGoBack} onclick={()=>move(-1)}/><IconButton name="right" label={text($locale,'Келесі тайтлдар','Следующие тайтлы')} disabled={!canGoForward} onclick={()=>move(1)}/></div>{/if}
</div>
<style>
 .shelf-wrap{position:relative;min-width:0}.cover-shelf{display:flex;gap:10px;overflow-x:auto;overscroll-behavior-x:contain;scroll-snap-type:x proximity;scrollbar-width:none;padding:4px 0 12px}.shelf-item{flex:0 0 116px;min-width:0;scroll-snap-align:start}.shelf-controls{position:absolute;top:68px;left:0;right:0;display:flex;justify-content:space-between;pointer-events:none}.shelf-controls :global(button){pointer-events:auto;background:var(--panel);border:1px solid var(--muted);border-radius:50%;box-shadow:0 2px 8px #0004}.shelf-controls :global(button:disabled){visibility:hidden}@media(min-width:768px){.cover-shelf{gap:10px}.shelf-item{flex-basis:146px}.shelf-controls{top:84px}}
</style>
