<script lang="ts">
 import {onMount} from 'svelte';
 import type {ReadingGateway,ProgressStorage} from '../application/progress';
 import {loadContinuation} from '../application/continuation';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import LibraryRow from './LibraryRow.svelte';
 let {readingGateway,progressStorage}:{readingGateway:ReadingGateway;progressStorage:ProgressStorage}=$props();
 const {locale}=useLocale();
 let rows=$state<{titleId:string;name:string;slug:string;releaseId:string}[]>([]);
 let valid=true;
 onMount(()=>{
  const stop=progressStorage.watchAccountChange(()=>{valid=false;rows=[];});
  void(async()=>{try{
   const loaded=await loadContinuation(readingGateway,progressStorage);
   if(valid)rows=loaded;
  }catch{rows=[];}})();
  return()=>{valid=false;stop();};
 });
</script>
{#if rows.length}<section class="continue-reading"><SectionHeading title={text($locale,'Оқуды жалғастыр','Продолжить чтение')}>{#snippet actions()}<a class="auth-link" href="/library">{text($locale,'Кітапхана','Библиотека')}</a>{/snippet}</SectionHeading>{#each rows as item (item.titleId)}<LibraryRow name={item.name} href={'/titles/'+encodeURIComponent(item.slug)} releaseId={item.releaseId} continueLabel={text($locale,'Жалғастыру','Продолжить')}/>{/each}</section>{/if}
<style>.continue-reading { margin-bottom:32px; }</style>
