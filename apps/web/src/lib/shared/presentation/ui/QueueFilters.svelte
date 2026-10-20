<script lang="ts">
 import {page} from '$app/state';
 import {useLocale,text} from '../locale';
 let {kinds=[]}:{kinds?:{value:string;kk:string;ru:string}[]}=$props();
 const {locale}=useLocale();
</script>
<form method="GET" data-sveltekit-reload><label>{text($locale,'Команда: аты немесе ID','Команда: название или ID')}<input name="team" value={page.url.searchParams.get('team')||''} maxlength="160"/></label>{#if kinds.length}<label>{text($locale,'Жұмыс түрі','Тип работы')}<select name="kind" value={page.url.searchParams.get('kind')||''}><option value="">{text($locale,'Барлығы','Все')}</option>{#each kinds as kind}<option value={kind.value}>{text($locale,kind.kk,kind.ru)}</option>{/each}</select></label>{/if}<button class="button subtle">{text($locale,'Қолдану','Применить')}</button></form>
<style>form{display:grid;gap:12px;max-width:680px;margin:20px 0}label{display:grid;gap:8px}button{justify-self:start}</style>
