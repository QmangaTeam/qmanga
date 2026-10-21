<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {items,id}:{items:{key:string;href:string;label:string}[];id:string}=$props();
 const {locale}=useLocale();
 let expanded=$state(false),canExpand=$state(false);
 $effect(()=>{id;expanded=false;});
 function observe(node:HTMLDivElement){
  let disposed=false;
  const measure=()=>{if(disposed)return;const limit=parseFloat(getComputedStyle(node).getPropertyValue('--preview-height'));canExpand=node.scrollHeight>limit+1;if(!canExpand)expanded=false;for(const link of node.querySelectorAll('a'))link.toggleAttribute('inert',!expanded&&link.offsetTop+link.offsetHeight>limit+1);};
  const resize=new ResizeObserver(measure);resize.observe(node);
  const changes=new MutationObserver(measure);changes.observe(node,{childList:true,subtree:true,characterData:true});
  document.fonts.addEventListener('loadingdone',measure);void document.fonts.ready.then(measure);measure();
  return {destroy(){disposed=true;resize.disconnect();changes.disconnect();document.fonts.removeEventListener('loadingdone',measure);}};
 }
</script>
{#if items.length}<div class="title-taxonomy">
 <div {id} class="title-genres" class:expanded use:observe>{#each items as item (item.key)}<a href={item.href}>{item.label}</a>{/each}</div>
 {#if canExpand}<button class="tags-toggle" type="button" aria-expanded={expanded} aria-controls={id} onclick={()=>expanded=!expanded}>{expanded?text($locale,'Жию','Свернуть'):text($locale,'Толығырақ','Подробнее')}<Icon name={expanded?'up':'down'} size={16}/></button>{/if}
</div>{/if}
<style>
 .title-taxonomy{min-width:0;margin-top:10px}.title-genres{position:relative;--preview-height:94px;display:flex;flex-wrap:wrap;align-content:start;gap:6px;max-height:var(--preview-height);overflow:hidden;min-width:0}.title-genres.expanded{max-height:none}
 a{display:inline-flex;align-items:center;min-height:44px;max-width:100%;padding:4px 9px;border:1px solid var(--line);border-radius:6px;background:var(--panel);font-size:12px;font-weight:500;line-height:1.4;text-decoration:none;color:var(--text);overflow-wrap:anywhere}
 a:hover{border-color:var(--muted)}.tags-toggle{display:inline-flex;align-items:center;gap:6px;min-height:44px;padding:4px 0;border:0;background:none;color:var(--accent);font:inherit;font-size:13px;cursor:pointer}.tags-toggle:focus-visible,a:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
 @media(max-width:767px){.title-genres{position:relative;--preview-height:44px}}
</style>
