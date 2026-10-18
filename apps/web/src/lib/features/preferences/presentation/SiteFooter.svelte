<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {version='local'}:{version?:string}=$props();const {locale}=useLocale();
 const columns=$derived([
  {name:'Qmanga',links:[{href:'/',label:text($locale,'Басты бет','Главная')},{href:'/legal/terms',label:text($locale,'Пайдалану шарттары','Условия использования')},{href:'/legal/copyright',label:text($locale,'Авторлық құқық','Авторские права')},{href:'/legal/privacy',label:text($locale,'Жеке деректер','Персональные данные')}]},
  {name:text($locale,'Оқырмандарға','Читателям'),links:[{href:'/catalog',label:'Каталог'},{href:'/library',label:text($locale,'Кітапхана','Библиотека')},{href:'/saved',label:text($locale,'Жүктелгендер','Скачанное')}]},
  {name:text($locale,'Аудармашыларға','Переводчикам'),links:[{href:'/studio/teams',label:text($locale,'Команда','Команда')},{href:'/studio/requests',label:text($locale,'Жобаға өтінім','Заявка на проект')},{href:'/studio/projects',label:text($locale,'Тараулар','Главы')},{href:'/help#publishing',label:text($locale,'Жариялау және модерация','Публикация и модерация')}]},
  {name:text($locale,'Қосымша','Дополнительно'),links:[{href:'/settings',label:text($locale,'Баптаулар','Настройки')},{href:'/reports',label:text($locale,'Өтініштерім','Мои обращения')},{href:'/help',label:text($locale,'Көмек','Помощь')},{href:'/rights',label:text($locale,'Құқық иелеріне','Правообладателям')}]}
 ]);
</script>
<footer class="site-footer" aria-label={text($locale,'Сайт туралы','О сайте')}>
 <div class="footer-columns">{#each columns as column}<section><h2>{column.name}</h2><ul>{#each column.links as link}<li><a href={link.href}>{link.label}</a></li>{/each}</ul></section>{/each}</div>
 <div class="footer-meta"><p>© 2026 Qmanga · {text($locale,'Қазақ тіліндегі манга, манхва және маньхуа.','Манга, манхва и маньхуа на казахском языке.')}</p><p>{text($locale,'Жинақ','Сборка')}: <code>{version}</code></p></div>
</footer>
<style>
 .site-footer{display:block}.footer-columns{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:20px 16px;width:100%}h2{font-size:14px;font-weight:650;color:var(--text);margin:0 0 8px}ul{list-style:none;padding:0;margin:0}a{color:var(--muted);font-size:13px;line-height:1.4;text-decoration:none;overflow-wrap:anywhere}a:hover{color:var(--text)}.footer-meta{display:flex;align-items:start;justify-content:space-between;flex-wrap:wrap;gap:8px 20px;margin-top:24px;padding-top:16px;border-top:1px solid var(--line);font-size:12px;line-height:1.6}.footer-meta p{margin:0}code{font-size:11px;overflow-wrap:anywhere}@media(min-width:768px){.footer-columns{grid-template-columns:repeat(4,minmax(0,1fr));gap:24px}}
</style>
