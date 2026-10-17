<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import {reportCategories} from './messages';
 const categoryNames=$derived(reportCategories($locale));
 import {onMount} from 'svelte';
 import type {CommunityGateway,ModerationReport} from '../application/community';
 import {validReport} from '../domain/content';
 let {gateway}:{gateway:CommunityGateway}=$props();
 let account=$state<string|null>(null),reports=$state<ModerationReport[]>([]),reasons=$state<Record<string,string>>({}),busy=$state(false),message=$state(''),more=$state(false);
 async function load(after=''){busy=true;try{account=await gateway.account();if(!account){message="login_required";return;}const result=await gateway.reports(account,after);reports=after?[...reports,...result.items]:result.items;more=result.hasMore;}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 onMount(()=>{void load();});
 async function decide(item:ModerationReport,status:string){if(!account||!validReport(reasons[item.id]||'')){message="decision_reason";return;}busy=true;try{await gateway.resolve(account,item.id,status,reasons[item.id]);reports=reports.filter(report=>report.id!==item.id);message="decision_saved";}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 async function visibility(item:ModerationReport,hidden:boolean){if(!account||!validReport(reasons[item.id]||'')){message="action_reason";return;}busy=true;try{await gateway.visibility(account,item.targetId,hidden,reasons[item.id]);message=hidden?"comment_hidden":"comment_restored";}catch(failure){message=failureCode(failure);}finally{busy=false;}}
</script>
<svelte:head><title>{copy($locale,"Шағымдар — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head>
<a class="auth-link" href="/moderation">{copy($locale,"Жарияланымдарды тексеру")}</a><h1>{copy($locale,"Шағымдар")}</h1>{#if message}<p class="form-message" role="status">{noticeText($locale,message)}</p>{/if}
{#each reports as item}<article><h2>{categoryNames[item.category]||item.category}</h2><p>{item.reason}</p><dl><dt>{copy($locale,"Нысан")}</dt><dd>{item.targetType} · {item.targetId}</dd>{#if item.pageId}<dt>{copy($locale,"Бет")}</dt><dd>{item.pageId}</dd>{/if}</dl>{#if item.releaseId}<a class="auth-link" href={`/read/${item.releaseId}`}>{copy($locale,"Тарауды қарау")}</a>{/if}<label>{copy($locale,"Модератор шешімінің себебі")}<textarea bind:value={reasons[item.id]} minlength="8" maxlength="2000" rows="3"></textarea></label><div class="actions">{#if item.targetType==='comment'}<button class="button subtle" disabled={busy} onclick={()=>visibility(item,true)}>{copy($locale,"Пікірді жасыру")}</button><button class="button subtle" disabled={busy} onclick={()=>visibility(item,false)}>{copy($locale,"Қайта ашу")}</button>{/if}<button class="button primary" disabled={busy} onclick={()=>decide(item,'resolved')}>{copy($locale,"Шешілді")}</button><button class="button subtle" disabled={busy} onclick={()=>decide(item,'dismissed')}>{copy($locale,"Негізсіз")}</button></div></article>{:else}{#if !busy}<p class="empty-state">{copy($locale,"Қаралмаған шағым жоқ.")}</p>{/if}{/each}{#if more}<button class="button" disabled={busy} onclick={()=>load(reports[reports.length-1].id)}>{copy($locale,"Тағы жүктеу")}</button>{/if}{#if busy}<p role="status">{copy($locale,"Жүктелуде…")}</p>{/if}
<style>h1{font-size:26px}h2{font-size:20px}article{padding:24px 0;border-bottom:1px solid var(--line)}article p{white-space:pre-wrap;overflow-wrap:anywhere;margin:12px 0}dl{font-size:13px}dd{margin:0;overflow-wrap:anywhere}label{display:grid;gap:8px;font-size:14px}.actions{display:flex;flex-wrap:wrap;gap:10px;margin-top:12px}</style>
