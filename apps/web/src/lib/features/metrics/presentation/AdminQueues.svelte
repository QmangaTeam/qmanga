<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 const {locale}=useLocale();
 const queues=$derived([
 {href:'/admin/moderation',name:text($locale,'Жарияланымдарды тексеру','Проверка публикаций'),description:text($locale,'Жариялауға жіберілген тараулар','Главы, отправленные на публикацию')},
 {href:'/moderation/title-changes',name:text($locale,'Тайтл өзгерістері','Изменения тайтлов'),description:text($locale,'Деректер, мұқабалар және баннерлер','Данные, обложки и баннеры')},
 {href:'/admin/requests',name:text($locale,'Жоба өтінімдері','Заявки на проекты'),description:text($locale,'Командалардың аударма өтінімдері','Запросы команд на перевод')},
 {href:'/admin/teams',name:text($locale,'Команда өтінімдері','Заявки команд'),description:text($locale,'Тіркелген командаларды қарау','Рассмотрение зарегистрированных команд')},
 {href:'/admin/reports',name:text($locale,'Шағымдар','Жалобы'),description:text($locale,'Мазмұн және пікірлер туралы шағымдар','Жалобы на содержание и комментарии')}
 ]);
</script>
<section aria-labelledby="admin-queues"><h2 id="admin-queues">{text($locale,'Шешімді күтіп тұрған жұмыстар','Работа, ожидающая решения')}</h2><ul>{#each queues as queue}<li><a href={queue.href}>{queue.name}</a><span>{queue.description}</span></li>{/each}</ul></section>
<style>section{margin-bottom:28px}h2{font-size:20px}ul{list-style:none;margin:0;padding:0;display:grid;gap:0;max-width:900px}li{padding:8px 0;border-bottom:1px solid var(--line);min-width:0}a{display:flex;align-items:center;min-height:44px;font-weight:600;overflow-wrap:anywhere}span{display:block;color:var(--muted);font-size:14px;padding-bottom:8px;overflow-wrap:anywhere}</style>
