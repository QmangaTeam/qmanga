<script lang="ts">
 import {untrack} from 'svelte';
 import {provideLocale} from '$lib/shared/presentation/locale';
 import SiteFooter from '$lib/features/preferences/presentation/SiteFooter.svelte';
 let {language='kk'}:{language?:'kk'|'ru'}=$props();provideLocale(untrack(()=>language),()=>{});
</script>
<SiteFooter version="abcdef1"/>
