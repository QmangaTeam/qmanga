<script lang="ts">
 import type {CommunityGateway,Report} from '../application/community';
 import {categoryNames,validReport} from '../domain/content';
 let {gateway,titleId,releaseId='',pageId='',targetType='title',targetId}:{gateway:CommunityGateway;titleId:string;releaseId?:string;pageId?:string;targetType?:Report['targetType'];targetId:string}=$props();
 let category=$state('other'),reason=$state(''),busy=$state(false),message=$state('');
 async function send(){if(!validReport(reason)){message='Себеп 8–2000 таңбадан тұруы керек';return;}busy=true;try{const account=await gateway.account();if(!account){message='Шағым жіберу үшін кіріңіз және поштаңызды растаңыз';return;}await gateway.report(account,{titleId,releaseId,pageId,targetType,targetId,category,reason});reason='';message='Шағым модераторға жіберілді';}catch(failure){message=failure instanceof Error?failure.message:'Шағым жіберілмеді';}finally{busy=false;}}
</script>
<details><summary>Мәселе туралы хабарлау</summary><form onsubmit={event=>{event.preventDefault();void send();}}><label>Мәселе түрі<select bind:value={category}>{#each Object.entries(categoryNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label><label>Түсіндірме<textarea bind:value={reason} minlength="8" maxlength="2000" required rows="4" placeholder="Нақты мәселені сипаттаңыз"></textarea></label><button class="button subtle" disabled={busy}>Шағымды жіберу</button>{#if message}<p role="status">{message}</p>{/if}</form></details>
<style>details{margin-top:20px;font-size:14px}summary{min-height:44px;display:flex;align-items:center;cursor:pointer;color:var(--muted,#aaa)}form{display:grid;gap:12px;margin:12px 0}label{display:grid;gap:8px}p{font-size:13px;overflow-wrap:anywhere}</style>
