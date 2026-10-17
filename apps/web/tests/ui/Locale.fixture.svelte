<script lang="ts">
 import {provideLocale} from '$lib/shared/presentation/locale';
 import type {Locale} from '$lib/features/preferences/domain/locale';
 import Catalogue from '$lib/features/catalog/presentation/Catalogue.svelte';
 import Help from '$lib/features/preferences/presentation/Help.svelte';
 import {untrack} from 'svelte';
 let {language,screen='catalogue'}:{language:Locale;screen?:'catalogue'|'help'}=$props();
 provideLocale(untrack(()=>language),()=>{});
</script>
{#if screen==='help'}<Help/>{:else}<Catalogue view={{available:true,message:'',filter:{},genres:[],page:{items:[],nextCursor:null},nextURL:null}}/>{/if}
