<script lang="ts">
 import {untrack} from 'svelte';
 import {provideLocale} from '$lib/shared/presentation/locale';
 import type {Locale} from '$lib/features/preferences/domain/locale';
 import type {Account} from '$lib/features/identity/domain/account';
 import Menu from '$lib/features/preferences/presentation/Menu.svelte';
 let {language='kk',role='reader',guest=false,verified=true}:{language?:Locale;role?:Account['role'];guest?:boolean;verified?:boolean}=$props();
 provideLocale(untrack(()=>language),()=>{});
 const user=$derived(guest?null:{id:'account',email:'private@example.test',role,verified,displayName:'Оқырман'});
</script>
<Menu {user}/>
