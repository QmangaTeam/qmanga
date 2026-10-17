<script lang="ts">
 import {page} from '$app/stores';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import WorkspaceNavigation from '$lib/shared/presentation/ui/WorkspaceNavigation.svelte';
 let {children}=$props();const {locale}=useLocale();
 const items=$derived([{href:'/admin',label:text($locale,'Командалар','Команды')},{href:'/admin/requests',label:text($locale,'Жоба өтінімдері','Заявки на проекты')},{href:'/admin/catalog',label:'Каталог'},{href:'/moderation',label:text($locale,'Модерация','Модерация')},{href:'/moderation/accounts',label:text($locale,'Аккаунттар','Аккаунты')},{href:'/admin/audit',label:text($locale,'Журнал','Журнал')}]);
</script>
<WorkspaceNavigation {items} currentPath={$page.url.pathname} label={text($locale,'Басқару','Управление')}/>
{@render children()}
