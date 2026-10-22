<script lang="ts">
 import WorkspaceShell from '../../src/lib/shared/presentation/ui/WorkspaceShell.svelte';
 let {compact=false}=$props();
</script>
<WorkspaceShell items={[{href:'/studio/projects',label:'Проекты',icon:'folder'},{href:'/studio/teams',label:'Команды',icon:'users'}]} currentPath="/studio/projects/example" label="Студия" {compact} unreadCount={3} unreadBadge="3">
 {#snippet account()}<a href="/settings">Аккаунт</a>{/snippet}
 <h1>Тестовая рабочая область</h1>
</WorkspaceShell>
