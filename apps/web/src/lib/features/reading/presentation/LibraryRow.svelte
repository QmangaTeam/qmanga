<script lang="ts">
 import type { Snippet } from 'svelte';
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let { name, href, coverURL='', statusLabel='', detail, releaseId, continueLabel, children, progress }: {name:string;href:string;coverURL?:string;statusLabel?:string;detail?:string;releaseId?:string;continueLabel:string;children?:Snippet;progress?:Snippet} = $props();
 const {locale}=useLocale();
 let dialog:HTMLDialogElement,trigger:HTMLButtonElement;
 let failed=$state(false);
 const heading=$derived('library-options-'+encodeURIComponent(href));
 $effect(()=>{coverURL;failed=false;});
 function close(){dialog.close();trigger.focus();}
 function backdrop(event:MouseEvent){if(event.target!==dialog)return;const rect=dialog.getBoundingClientRect();if(event.clientX<rect.left||event.clientX>rect.right||event.clientY<rect.top||event.clientY>rect.bottom)close();}
</script>
<article class="library-card">
 <a class="cover" {href} tabindex="-1" aria-hidden="true">{#if coverURL&&!failed}<img src={coverURL} alt="" width="280" height="400" loading="lazy" onerror={()=>failed=true}/>{:else}<span>{name.slice(0,1)}</span>{/if}</a>
 <div class="card-heading"><a class="title" {href}><span>{name}</span></a><button bind:this={trigger} class="options" aria-label={name+': '+text($locale,'тізімді өзгерту','изменить закладку')} aria-haspopup="dialog" onclick={()=>dialog.showModal()}>⋮</button></div>
 {#if statusLabel}<p class="library-status">{statusLabel}</p>{/if}
 {#if progress}{@render progress()}{/if}
 <dialog bind:this={dialog} aria-labelledby={heading} onclick={backdrop}>
  <header><h2 id={heading}>{name}</h2><button class="close" aria-label={text($locale,'Жабу','Закрыть')} onclick={close}><Icon name="close"/></button></header>
  {#if detail}<p class="muted">{detail}</p>{/if}
  {#if releaseId}<ActionLink href={'/read/'+encodeURIComponent(releaseId)} variant="primary">{continueLabel}</ActionLink>{/if}
  {#if children}<div class="row-controls">{@render children()}</div>{/if}
 </dialog>
</article>
<style>
 .library-status{margin:2px 0 0;font-size:12px;line-height:18px;color:var(--muted);overflow-wrap:anywhere}article{min-width:0}.cover{display:grid;place-items:center;aspect-ratio:7/10;border-radius:10px;overflow:hidden;background:var(--raised);color:var(--muted);text-decoration:none}.cover img{display:block;width:100%;height:100%;object-fit:cover}.cover span{font-size:32px}.card-heading{display:flex;align-items:start;gap:2px;min-width:0;margin-top:6px}.title{min-width:0;flex:1;min-height:44px;padding:4px 0;color:var(--text);font-size:12px;line-height:17px;font-weight:600;text-decoration:none;overflow-wrap:anywhere;display:block}.title span{display:-webkit-box;-webkit-box-orient:vertical;-webkit-line-clamp:2;line-clamp:2;overflow:hidden;max-height:34px}.options,.close{height:44px;flex:none;border:0;border-radius:8px;background:transparent;color:var(--text);cursor:pointer}.options{width:44px;font-size:25px}.options:hover,.close:hover{background:var(--raised)}dialog{width:min(480px,calc(100vw - 24px));max-height:calc(100dvh - 32px);padding:20px;border:1px solid var(--line);border-radius:16px;background:var(--panel);color:var(--text);overflow:auto}dialog::backdrop{background:#0009}header{display:flex;align-items:start;gap:12px;margin-bottom:16px}h2{flex:1;min-width:0;margin:0;font-size:19px;overflow-wrap:anywhere}.close{width:44px;display:grid;place-items:center}.row-controls{display:grid;gap:16px;margin-top:16px}p{font-size:14px}
 @media(min-width:768px){.title{font-size:15px;line-height:21px}.title span{max-height:42px}.options{width:44px}}
</style>
