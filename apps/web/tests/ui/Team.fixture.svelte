<script lang="ts">
 import {provideLocale} from '$lib/shared/presentation/locale';
 import PublicTeam from '$lib/features/catalog/presentation/PublicTeam.svelte';
 import type {PublicTeam as Team} from '$lib/features/catalog/application/catalog';
 import type {Locale} from '$lib/features/preferences/domain/locale';
 import {untrack} from 'svelte';
 let {language,team}:{language:Locale;team:Team}=$props();
 provideLocale(untrack(()=>language),()=>{});
</script>
<PublicTeam {team}/>
