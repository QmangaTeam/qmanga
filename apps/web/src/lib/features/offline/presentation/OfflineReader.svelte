<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import {onMount} from 'svelte';
 import Reader from '../../reading/presentation/Reader.svelte';
 import type {Download,Downloads} from '../application/downloads';
 import type {ImageLoader,ReaderGateway,ReaderPlatform,Timer} from '../../reading/application/reader';
 import type {CatalogGateway} from '../../catalog/application/catalog';
 import type {ReadingGateway,ProgressStorage} from '../../reading/application/progress';
 let {downloads,ports,releaseId}:{releaseId?:string;downloads:Downloads;ports:{gateway:ReaderGateway;catalogue:CatalogGateway;platform:ReaderPlatform;images:(first?:HTMLImageElement)=>ImageLoader;timer:Timer;readingGateway?:ReadingGateway;progressStorage?:ProgressStorage}}=$props();
 let selected=$state<Download|null>(null),message=$state("offline_checking");
 onMount(()=>{void (async()=>{try{const release=releaseId;const list=(await downloads.list()).filter(item=>!release||item.manifest.releaseId===release).sort((left,right)=>right.updatedAt-left.updatedAt);for(const item of list){const reconciled=await downloads.reconcile(item);if(reconciled.status==='ready'){selected=reconciled;return;}}message="offline_missing";}catch{message="offline_storage";}})();});
</script>
{#if selected}<Reader manifest={selected.manifest} offlineTier={selected.tier} {...ports}/>{:else}<div class="offline-empty"><h1>{copy($locale,"Офлайн оқу")}</h1><p role="status">{noticeText($locale,message)}</p><a class="button" href="/saved" data-sveltekit-reload>{copy($locale,"Сақталған тараулар")}</a></div>{/if}
<style>.offline-empty{max-width:560px;margin:80px auto;padding:24px}.offline-empty h1{font-size:24px}.offline-empty p{margin:20px 0}</style>
