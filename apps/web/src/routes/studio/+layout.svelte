<script lang="ts">
 import {page} from '$app/stores';
 import WorkspaceNavigation from '$lib/shared/presentation/ui/WorkspaceNavigation.svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {children}=$props();const {locale}=useLocale();
 const items=$derived([{href:'/studio/projects',label:text($locale,'Жобалар','Проекты')},{href:'/studio/requests',label:text($locale,'Өтінімдер','Заявки')},{href:'/studio/teams',label:text($locale,'Командалар','Команды')},{href:'/studio/stats',label:text($locale,'Статистика','Статистика')}]);
</script>
<div class="studio-workspace"><aside><WorkspaceNavigation sidebar menuLabel={text($locale,'Бөлімдер','Разделы')} {items} currentPath={$page.url.pathname} label={text($locale,'Аудармашы кабинеті','Кабинет переводчика')}/></aside>
<div class="workspace-content">{@render children()}</div></div>
<style>
 .studio-workspace{display:grid;gap:24px;min-width:0}.workspace-content,aside{min-width:0}
 @media(min-width:900px){.studio-workspace{grid-template-columns:200px minmax(0,1fr);gap:32px}}
</style>
