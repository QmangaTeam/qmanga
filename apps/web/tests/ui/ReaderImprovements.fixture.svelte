<script lang="ts">
 import {provideLocale} from '$lib/shared/presentation/locale';
 import LibraryRow from '$lib/features/reading/presentation/LibraryRow.svelte';
 import Catalogue from '$lib/features/catalog/presentation/Catalogue.svelte';
 let {screen='library'}:{screen?:string}=$props();
 provideLocale('ru',()=>{});
</script>
{#if screen==='library'}<LibraryRow name="Тайтл" href="/titles/title" continueLabel="Продолжить" statusLabel="Прочитано"/>{:else}<Catalogue view={{available:false,message:'',filter:{query:'query',genres:['action'],sort:'name'},genres:[],page:{items:[],nextCursor:null},nextURL:null}}/>{/if}
