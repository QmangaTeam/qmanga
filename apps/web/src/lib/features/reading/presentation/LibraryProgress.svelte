<script lang="ts">
 import {onMount} from 'svelte';
 import type {LibraryEntry} from '../application/progress';
 import type {LibraryProgressReader,LibraryPosition} from '../application/library-progress';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {revealFill} from '$lib/shared/presentation/ui/reveal-fill';
 let {entry,reader}:{entry:LibraryEntry;reader:LibraryProgressReader}=$props();
 const {locale}=useLocale();
 let node:HTMLDivElement,visible=$state(false),value=$state<LibraryPosition|null>(null),loading=$state(false),alive=false,generation=0;
 onMount(()=>{alive=true;if(typeof IntersectionObserver==='undefined')visible=true;else{const observer=new IntersectionObserver(entries=>{if(entries.some(item=>item.isIntersecting)){visible=true;observer.disconnect();}});observer.observe(node);return()=>{alive=false;generation++;observer.disconnect();};}return()=>{alive=false;generation++;};});
 $effect(()=>{const item=entry,run=++generation;value=null;if(!visible||!item.progress.releaseId){loading=false;return;}loading=true;void reader.read(item).then(result=>{if(alive&&run===generation){value=result;loading=false;}});return()=>{generation++;};});
</script>
<div class="library-progress" bind:this={node}>
 {#if value}<progress value={value.fraction} max="1" aria-label={text($locale,'Тараудағы оқу барысы','Прогресс чтения главы')}>{Math.round(value.fraction*100)}%</progress><span class="track motion-track" use:revealFill aria-hidden="true"><span class="motion-fill" style:width={value.fraction*100+'%'}></span></span><p><span>{text($locale,value.chapter+'-тарау · '+value.page+' / '+value.pages+' бет','Гл. '+value.chapter+' · стр. '+value.page+' / '+value.pages)}</span><span>{Math.round(value.fraction*100)}%</span></p>
 {:else}<p class="state">{!entry.progress.releaseId?text($locale,'Әлі басталмаған','Ещё не начато'):loading?text($locale,'Оқу орны жүктелуде…','Загрузка прогресса…'):text($locale,'Оқу орны қолжетімсіз','Прогресс недоступен')}</p>{/if}
</div>
<style>
 .library-progress{position:relative;margin-top:8px;min-width:0}progress{position:absolute;width:1px;height:1px;clip-path:inset(50%);overflow:hidden}.track{display:block;height:4px;border-radius:3px;overflow:hidden;background:var(--line)}.track>span{display:block;height:100%;border-radius:inherit;background:var(--accent)}p{display:flex;justify-content:space-between;gap:6px;font-size:10px;line-height:1.5;color:var(--muted);margin:6px 0 0;overflow-wrap:anywhere}p>span:last-child{flex:none}.state{display:block}
</style>
