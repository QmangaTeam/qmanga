<script lang="ts">
 import {onMount,untrack} from 'svelte';
 import type {ReadingGateway,ProgressStorage,ReadingChanges} from '../application/progress';
 import {RegistrationImport,type RegistrationPermit} from '../application/registration-import';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {readingGateway,progressStorage,registrationPermit,readingChanges}:{readingChanges:ReadingChanges;readingGateway:ReadingGateway;progressStorage:ProgressStorage;registrationPermit:RegistrationPermit}=$props();
 const {locale}=useLocale();let version=$state(0);
 const importer=new RegistrationImport(untrack(()=>readingGateway),untrack(()=>progressStorage),untrack(()=>registrationPermit),()=>version++,()=>readingChanges.publish());
 const failed=$derived.by(()=>{version;return importer.failed;});
 async function attempt(){await importer.run();}
 onMount(()=>{
  void attempt();const changed=()=>{importer.clear();void attempt();};const focus=()=>{if(document.visibilityState==='visible')void attempt();};
  let stop=()=>{};try{stop=progressStorage.watchAccountChange(changed);}catch{/* The import attempt handles unavailable local storage. */}window.addEventListener('focus',focus);
  return()=>{stop();window.removeEventListener('focus',focus);importer.close();};
 });
</script>
{#if failed}<p role="status" class="import-notice">{text($locale,'Қонақтағы оқу орнының бір бөлігі әлі көшірілмеді. Ол осы құрылғыда сақталған; кейінірек қайталап көреміз.','Часть гостевого чтения пока не перенеслась. Она сохранена на этом устройстве; попробуем позже.')}</p>{/if}
<style>.import-notice{margin:12px 16px;padding:12px 16px;border:1px solid var(--line);border-radius:10px;background:var(--panel);color:var(--muted);font-size:14px;line-height:1.5}</style>
