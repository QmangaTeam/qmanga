<script lang="ts">
 import {page} from '$app/stores';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import WorkspaceNavigation from '$lib/shared/presentation/ui/WorkspaceNavigation.svelte';
 let {children,data}=$props();const {locale}=useLocale();
 const items=$derived([{href:'/moderation',label:text($locale,'Жарияланымдар','Публикации')},{href:'/moderation/reports',label:text($locale,'Шағымдар','Обращения')},{href:'/moderation/accounts',label:text($locale,'Аккаунттар','Аккаунты')},...(data.user?.role==='admin'?[{href:'/admin/requests',label:text($locale,'Жоба өтінімдері','Заявки на проекты')},{href:'/admin',label:text($locale,'Командалар','Команды')}]:[])]);
</script>
<WorkspaceNavigation {items} currentPath={$page.url.pathname} label={text($locale,'Модерация','Модерация')}/>
{@render children()}
