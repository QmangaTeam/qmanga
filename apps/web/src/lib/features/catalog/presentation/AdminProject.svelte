<script lang="ts">
 import { onMount } from 'svelte';
 import type { Account } from '$lib/features/identity/domain/account';
 import type { Project } from '../application/catalog';
 import type { CatalogManagementGateway } from '../application/management';
 import { useLocale, text } from '$lib/shared/presentation/locale';
 import { failureCode, noticeText } from '$lib/shared/presentation/copy';

 let { user, projectId, gateway }: { user: Account | null; projectId: string; gateway: CatalogManagementGateway } = $props();
 const { locale } = useLocale();
 let project = $state<Project | null>(null);
 let busy = $state(false);
 let failure = $state('');
 async function load() {
  if (busy || user?.role !== 'admin') return;
  busy = true;
  failure = '';
  try { project = await gateway.project(projectId); }
  catch (error) { failure = failureCode(error); }
  finally { busy = false; }
 }
 onMount(load);
</script>

<svelte:head><title>{text($locale, 'Жоба — Qmanga', 'Проект — Qmanga')}</title><meta name="robots" content="noindex"/></svelte:head>
<a class="button subtle" href="/admin/catalog">← {text($locale, 'Каталогқа', 'К каталогу')}</a>
<h1>{text($locale, 'Жоба', 'Проект')}</h1>
{#if user?.role !== 'admin'}
 <p role="alert">{text($locale, 'Әкімші рұқсаты қажет.', 'Нужны права администратора.')}</p>
{:else if failure}
 <p class="form-message" role="alert">{noticeText($locale, failure)}</p>
 <button class="button subtle" onclick={load}>{text($locale, 'Қайталау', 'Повторить')}</button>
{:else if !project}
 <p role="status">{text($locale, 'Жүктелуде…', 'Загрузка…')}</p>
{:else}
 <section class="panel" data-testid="admin-project">
  <h2>{project.titleName}</h2>
  <dl>
   <dt>{text($locale, 'Команда', 'Команда')}</dt><dd>{project.teamName}</dd>
   <dt>{text($locale, 'Күйі', 'Статус')}</dt><dd>{project.status === 'active' ? text($locale, 'Белсенді', 'Активен') : text($locale, 'Тоқтатылған', 'Приостановлен')}</dd>
  </dl>
  <p class="muted">{text($locale, 'Бұл — әкімшіге арналған шолу. Тарауларды команда дайындайды, жариялауды модератор тексереді.', 'Это административный просмотр. Главы готовит команда, публикацию проверяет модератор.')}</p>
  <div class="actions">
   <a class="button primary" href={'/admin/catalog?title=' + encodeURIComponent(project.titleId)}>{text($locale, 'Тайтлды өңдеу', 'Редактировать тайтл')}</a>
   <a class="button subtle" href="/moderation">{text($locale, 'Жарияланымдарды тексеру', 'Проверка публикаций')}</a>
  </div>
 </section>
{/if}

<style>
 h1 { margin-block: 20px; }
 h2, dd { overflow-wrap: anywhere; }
 dt { color: var(--muted); font-size: 14px; }
 dd { margin: 4px 0 16px; }
 .actions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 20px; }
</style>
