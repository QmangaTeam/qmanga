<script lang="ts">
 import {page} from '$app/stores';
 import WorkspaceNavigation from '$lib/shared/presentation/ui/WorkspaceNavigation.svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {children}=$props();const {locale}=useLocale();
 const items=$derived([{href:'/studio/projects',label:text($locale,'Жобалар','Проекты')},{href:'/studio/requests',label:text($locale,'Өтінімдер','Заявки')},{href:'/studio/teams',label:text($locale,'Командалар','Команды')}]);
</script>
<WorkspaceNavigation {items} currentPath={$page.url.pathname} label={text($locale,'Аудармашы кабинеті','Кабинет переводчика')}/>
{@render children()}
