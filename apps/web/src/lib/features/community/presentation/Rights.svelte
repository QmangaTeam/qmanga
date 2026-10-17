<script lang="ts">
 import {onMount} from 'svelte';
 import type {RightsGateway} from '../application/rights';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 let {gateway,initialTitle=''}:{gateway:RightsGateway;initialTitle?:string}=$props();
 const {locale}=useLocale();
 let titleSlug=$state(''),email=$state(''),reason=$state(''),id=$state(''),token=$state(''),busy=$state(false),message=$state(''),created=$state(false),result=$state<{status:string;resolution:string}|null>(null);
 onMount(()=>{titleSlug=initialTitle;});
 async function send(){busy=true;message='';try{const receipt=await gateway.create({titleSlug,email,reason});id=receipt.id;token=receipt.token;created=true;result=null;reason='';}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 async function check(){busy=true;message='';try{result=await gateway.status(id.trim(),token.trim());}catch(failure){result=null;message=failureCode(failure);}finally{busy=false;}}
 const status=(value:string)=>value==='pending'?text($locale,'Қаралуда','На рассмотрении'):value==='resolved'?text($locale,'Шешілді','Решено'):text($locale,'Қабылданбады','Отклонено');
</script>
<svelte:head><title>{text($locale,'Құқық иелеріне','Правообладателям')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={text($locale,'Құқық иелеріне','Правообладателям')} description={text($locale,'Аккаунтсыз өтініш жіберуге болады. Мәтін мен пошта тек модераторларға көрінеді.','Обращение можно отправить без аккаунта. Текст и почта видны только модераторам.')}/>
<div class="rights">
 <form onsubmit={event=>{event.preventDefault();void send();}}>
 <label>{text($locale,'Тайтлдың сілтемедегі атауы','Код тайтла из адреса')}<input bind:value={titleSlug} required maxlength="128" pattern="[a-z0-9-]+" placeholder="title-slug"/><small>{text($locale,'/titles/ сөзінен кейінгі бөлік','Часть адреса после /titles/')}</small></label>
 <label>{text($locale,'Байланыс поштасы','Почта для связи')}<input type="email" bind:value={email} required maxlength="254" autocomplete="email"/></label>
 <label>{text($locale,'Құқықтың негізі және өтініш','Основание прав и обращение')}<textarea bind:value={reason} required minlength="8" maxlength="2000" rows="6"></textarea></label>
 <p class="muted">{text($locale,'Құпия құжаттар мен құпиясөздерді жіберме. Пошта бұл формада расталмайды; өтініш автоматты түрде материалды өшірмейді.','Не присылай конфиденциальные документы и пароли. Почта в этой форме не подтверждается; обращение не удаляет материал автоматически.')}</p>
 <button class="button primary" disabled={busy}>{text($locale,'Өтініш жіберу','Отправить обращение')}</button></form>
 {#if created}<StateMessage kind="success" title={text($locale,'Өтініш қабылданды','Обращение принято')} description={text($locale,'Төмендегі нөмір мен құпия кодты сақта. Оларсыз нәтижені көру мүмкін емес. Кодты жариялама.','Сохрани номер и секретный код ниже. Без них нельзя посмотреть результат. Не публикуй код.')}/>{/if}
 <form onsubmit={event=>{event.preventDefault();void check();}}><h2>{text($locale,'Нәтижені тексеру','Проверить результат')}</h2>
 <label>{text($locale,'Өтініш нөмірі','Номер обращения')}<input bind:value={id} required maxlength="36" autocomplete="off"/></label>
 <label>{text($locale,'Құпия код','Секретный код')}<textarea bind:value={token} required maxlength="72" rows="3" autocomplete="off"></textarea></label>
 <button class="button subtle" disabled={busy}>{text($locale,'Тексеру','Проверить')}</button></form>
 {#if result}<StateMessage kind="success" title={status(result.status)} description={result.resolution||text($locale,'Жауап әлі жоқ.','Ответа пока нет.')}/>{/if}
 {#if message}<StateMessage kind="error" title={noticeText($locale,message)}/>{/if}
</div>
<style>.rights{max-width:680px}form{display:grid;gap:16px;padding:24px 0;border-bottom:1px solid var(--line)}label{display:grid;gap:8px;font-size:14px}small,.muted{color:var(--muted)}button{justify-self:start}textarea{overflow-wrap:anywhere}</style>
