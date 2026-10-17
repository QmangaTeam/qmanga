<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import {reportCategories} from './messages';
 const categoryNames=$derived(reportCategories($locale));
 import type {CommunityGateway,Report} from '../application/community';
 import {validReport} from '../domain/content';
 let {gateway,titleId,releaseId='',pageId='',targetType='title',targetId}:{gateway:CommunityGateway;titleId:string;releaseId?:string;pageId?:string;targetType?:Report['targetType'];targetId:string}=$props();
 let category=$state('other'),reason=$state(''),busy=$state(false),message=$state('');
 async function send(){if(!validReport(reason)){message="reason_length";return;}busy=true;try{const account=await gateway.account();if(!account){message="report_auth";return;}await gateway.report(account,{titleId,releaseId,pageId,targetType,targetId,category,reason});reason='';message="report_sent";}catch(failure){message=failureCode(failure);}finally{busy=false;}}
</script>
<details><summary>{copy($locale,"Мәселе туралы хабарлау")}</summary><form onsubmit={event=>{event.preventDefault();void send();}}><label>{copy($locale,"Мәселе түрі")}<select bind:value={category}>{#each Object.entries(categoryNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label><label>{copy($locale,"Түсіндірме")}<textarea bind:value={reason} minlength="8" maxlength="2000" required rows="4" placeholder={copy($locale,"Нақты мәселені сипаттаңыз")}></textarea></label><button class="button subtle" disabled={busy}>{copy($locale,"Шағымды жіберу")}</button>{#if message}<p role="status">{noticeText($locale,message)}</p>{/if}</form><a class="auth-link" href="/reports">{text($locale,'Өтініштерім және жауаптар','Мои обращения и ответы')}</a> · <a class="auth-link" href="/rights">{text($locale,'Құқық иелеріне — аккаунтсыз','Правообладателям — без аккаунта')}</a></details>
<style>details{margin-top:20px;font-size:14px}summary{min-height:44px;display:flex;align-items:center;cursor:pointer;color:var(--muted,#aaa)}form{display:grid;gap:12px;margin:12px 0}label{display:grid;gap:8px}p{font-size:13px;overflow-wrap:anywhere}</style>
