<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {status}:{status:string}=$props();const {locale}=useLocale();
 const labels=$derived<Record<string,string>>({pending:text($locale,'Қаралуда','На рассмотрении'),approved:text($locale,'Мақұлданды','Одобрено'),rejected:text($locale,'Қайтарылды','Возвращено'),cancelled:text($locale,'Бас тартылды','Отменено')});
</script>
<span class:approved={status==='approved'} class:rejected={status==='rejected'}>{labels[status]||status}</span>
<style>span{display:inline-flex;align-items:center;min-height:28px;padding:2px 8px;font-size:12px;border:1px solid var(--line);border-radius:4px;color:var(--muted)}.approved{color:var(--accent)}.rejected{color:var(--text);border-style:dashed}</style>
