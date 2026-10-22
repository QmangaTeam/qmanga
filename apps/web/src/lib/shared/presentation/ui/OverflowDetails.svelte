<script lang="ts">
 import type {Snippet} from 'svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from './Icon.svelte';
 let {id,label,children,previewHeight=240}:{id:string;label:string;children:Snippet;previewHeight?:number}=$props();
 const {locale}=useLocale();
 let expanded=$state(false),overflow=$state(false);
 $effect(()=>{id;expanded=false;});
 function observe(node:HTMLDivElement,options:{expanded:boolean;limit:number}){
  let current=options,disposed=false;
  const measure=()=>{
   if(disposed)return;
   overflow=node.scrollHeight>current.limit+1;
   if(!overflow)expanded=false;
   const top=node.getBoundingClientRect().top;
   for(const control of node.querySelectorAll<HTMLElement>('a,button,input,select,textarea,[tabindex]')){
    const rect=control.getBoundingClientRect(),container=control.matches('[tabindex]')&&control.querySelector('a,button,input,select,textarea');
    control.toggleAttribute('inert',!current.expanded&&(container?rect.top-top>=current.limit:rect.bottom-top>current.limit));
   }
  };
  const resize=new ResizeObserver(measure);resize.observe(node);
  document.fonts.addEventListener('loadingdone',measure);void document.fonts.ready.then(measure);measure();
  return {update(value:typeof options){current=value;measure();},destroy(){disposed=true;resize.disconnect();document.fonts.removeEventListener('loadingdone',measure);}};
 }
</script>
<div class="overflow-details">
 <div class="preview" class:expanded style:--preview-height={previewHeight+'px'}><div {id} use:observe={{expanded,limit:previewHeight}}>{@render children()}</div></div>
 {#if overflow}<button type="button" aria-controls={id} aria-expanded={expanded} aria-label={(expanded?text($locale,'Жию','Свернуть'):text($locale,'Толығырақ','Подробнее'))+': '+label} onclick={()=>expanded=!expanded}>{expanded?text($locale,'Жию','Свернуть'):text($locale,'Толығырақ','Подробнее')}<Icon name={expanded?'up':'down'} size={16}/></button>{/if}
</div>
<style>
 .preview{max-height:var(--preview-height);overflow:hidden}.preview.expanded{max-height:none}
 button{display:inline-flex;align-items:center;gap:6px;min-height:44px;padding:4px 0;border:0;background:none;color:var(--accent);font:inherit;font-size:13px;cursor:pointer}
 button:hover{text-decoration:underline}button:focus-visible{outline:2px solid var(--accent);outline-offset:3px}
</style>
