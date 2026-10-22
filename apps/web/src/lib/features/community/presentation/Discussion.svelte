<script lang="ts">
 import {replyCursor,mergeReplyPage} from '../application/reply-page';
 import {safeQueueReturn} from '$lib/shared/presentation/queue-context';
 import {page} from '$app/state';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 let emojiOpen=$state(false);
 const {locale,change:changeLanguage}=useLocale();
 import type {CommunityGateway,Comment} from '../application/community';
 import {validComment} from '../domain/content';
 import {onMount,tick} from 'svelte';
 import {invalidateAll} from '$app/navigation';
 import ReportForm from './ReportForm.svelte';
 import CommentBody from './CommentBody.svelte';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {titleId,titleSlug='',releaseId='',gateway,showReport=true}:{titleId:string;titleSlug?:string;releaseId?:string;gateway:CommunityGateway;showReport?:boolean}=$props();
 let opened=$state(false),loadFailed=$state(false),comments=$state<Comment[]>([]),next=$state(''),account=$state<string|null>(null),body=$state(''),spoiler=$state(false),parent=$state(''),editing=$state(''),reporting=$state(''),busy=$state(false),message=$state('');
 const focused=$derived(page.url.searchParams.get('comment')||'');
 let targetMissing=$state(false);
 let editor=$state<HTMLTextAreaElement|undefined>();
 let missingAvatars=$state<string[]>([]);
 let rootOpen=$state(false),rootDraft=$state({body:'',spoiler:false}),expandedRoots=$state<string[]>([]);
 const rootEditorId=$derived('root-editor-'+titleId+'-'+(releaseId||'title'));
 const threadId=(id:string)=>'comment-thread-'+id;
 const replyCount=(root:Comment)=>Math.max(root.replies,comments.filter(item=>item.rootId===root.id).length);
 function revealThread(id:string){if(!expandedRoots.includes(id))expandedRoots=[...expandedRoots,id];}
 async function toggleThread(root:Comment){if(expandedRoots.includes(root.id)){expandedRoots=expandedRoots.filter(id=>id!==root.id);return;}revealThread(root.id);if(!comments.some(item=>item.rootId===root.id)&&root.replies>0)await replies(root);}
 function rememberRoot(){if(!parent&&!editing)rootDraft={body,spoiler};}
 function openRoot(trigger:HTMLButtonElement){composerTrigger=trigger;body=rootDraft.body;spoiler=rootDraft.spoiler;rootOpen=true;void focusEditor();}
 let rootTrigger=$state<HTMLButtonElement|undefined>();
 let composerTrigger:HTMLButtonElement|undefined;
 async function cancelComposer(){const inline=!!(parent||editing);if(!inline){rememberRoot();rootOpen=false;}editing='';parent='';body=rootDraft.body;spoiler=rootDraft.spoiler;await tick();(inline?composerTrigger:rootTrigger)?.focus({preventScroll:true});}
 onMount(()=>{void load();});
 async function load(after=''){opened=true;busy=true;loadFailed=false;message='';try{const result=await gateway.discussion(titleId,releaseId,after,after?'':focused);comments=after?[...comments,...result.comments]:result.comments;next=result.next;account=await gateway.account();if(!after&&focused){targetMissing=!comments.some(item=>item.id===focused);const target=comments.find(item=>item.id===focused);if(target?.rootId)revealThread(target.rootId);await tick();document.getElementById('comment-'+focused)?.focus();document.getElementById('comment-'+focused)?.scrollIntoView({block:'center'});}}catch(failure){loadFailed=true;message=failureCode(failure);}finally{busy=false;}}
 async function focusEditor(){await tick();editor?.focus({preventScroll:true});}
 function replyTo(item:Comment,trigger:HTMLButtonElement){rememberRoot();revealThread(item.rootId||item.id);composerTrigger=trigger;parent=item.id;editing='';body='';spoiler=false;focusEditor();}
 function editComment(item:Comment,trigger:HTMLButtonElement){rememberRoot();if(item.rootId)revealThread(item.rootId);composerTrigger=trigger;editing=item.id;body=item.body;spoiler=item.spoiler;parent='';focusEditor();}

 async function submit(){const inline=!!(parent||editing);if(!account||!validComment(body)){message="comment_length";return;}busy=true;try{if(editing)await gateway.edit(account,editing,body,spoiler);else await gateway.comment(account,titleId,releaseId,parent,body,spoiler);if(!parent&&!editing)rootDraft={body:'',spoiler:false};if(parent)revealThread(comments.find(item=>item.id===parent)?.rootId||parent);body=rootDraft.body;parent='';editing='';spoiler=rootDraft.spoiler;rootOpen=false;await load();await invalidateAll();await tick();(inline?composerTrigger:rootTrigger)?.focus({preventScroll:true});}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 async function remove(id:string){if(!account)return;busy=true;try{await gateway.remove(account,id);await load();await invalidateAll();}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 async function replies(root:Comment){busy=true;try{const result=await gateway.replies(root.id,replyCursor(root,comments));comments=mergeReplyPage(comments,root,result.comments);}catch{message="replies_failed";}finally{busy=false;}}
 const author=(item:Comment)=>item.displayName||copy($locale,"Оқырман · ")+item.userId.slice(-6);
</script>
{#snippet avatar(item:Comment)}
 <span class="avatar" aria-hidden="true">
  {#if missingAvatars.includes(item.userId)}<span>{author(item).trim().slice(0,1).toLocaleUpperCase($locale)}</span>
  {:else}<img src={'/media/avatars/user/'+encodeURIComponent(item.userId)} alt="" width="36" height="36" loading="lazy" onerror={()=>missingAvatars=[...missingAvatars,item.userId]}/>{/if}
 </span>
{/snippet}
{#snippet composer()}
 <form class="comment-composer" onsubmit={event=>{event.preventDefault();void submit();}}>
  <div class="composer-context" class:root-context={!parent&&!editing}>
   <p class:sr-only={!parent&&!editing} title={parent?author(comments.find(item=>item.id===parent)!):undefined}>{#if parent}{text($locale,'Жауап: ','Ответ для ')}<strong>{author(comments.find(item=>item.id===parent)!)}</strong>{:else if editing}{copy($locale,"Пікірді өңдеу")}{:else}{copy($locale,"Пікір жазу")}{/if}</p>
   <button class="close-composer" type="button" disabled={busy} aria-label={text($locale,'Өңдегішті жабу','Закрыть редактор')} onclick={cancelComposer}>×</button>
  </div>
  <label><span class="sr-only">{editing?copy($locale,"Пікірді өңдеу"):parent?copy($locale,"Жауап жазу"):copy($locale,"Пікір жазу")}</span><textarea disabled={busy} bind:this={editor} bind:value={body} required maxlength="2000" rows="3" placeholder={editing?'':parent?text($locale,'Жауап жазыңыз…','Напишите ответ…'):text($locale,'Пікір жазыңыз…','Напишите комментарий…')}></textarea></label>
  {#if spoiler}<p class="spoiler-notice" role="status">{text($locale,'Спойлер қосулы — пікір жасырылып көрсетіледі.','Спойлер включён — комментарий будет скрыт.')}</p>{/if}
  <div class="editor-footer">
   <div class="format-actions" aria-label={text($locale,'Мәтін пішімі','Форматирование')}>
    <button disabled={busy} type="button" title={text($locale,'Спойлер','Спойлер')} aria-label={text($locale,'Спойлер','Спойлер')} aria-pressed={spoiler} onclick={()=>spoiler=!spoiler}><Icon name="eye-off"/></button>
    <div class="emoji-picker"><button type="button" disabled={busy} aria-expanded={emojiOpen} onclick={()=>emojiOpen=!emojiOpen} aria-label={text($locale,'Эмодзи қосу','Добавить эмодзи')} title={text($locale,'Эмодзи','Эмодзи')}><Icon name="smile"/></button>{#if emojiOpen}<div class="emoji-options" aria-label={text($locale,'Эмодзи','Эмодзи')}>{#each ['😊','❤️','👍','🔥','👏','😂','😮','🎉'] as emoji}<button type="button" disabled={busy} aria-label={text($locale,'Эмодзи қосу: ','Добавить эмодзи: ')+emoji} onclick={async event=>{const position=editor?.selectionStart??body.length;if(body.length+emoji.length>2000)return;body=body.slice(0,position)+emoji+body.slice(position);emojiOpen=false;await tick();editor?.focus({preventScroll:true});editor?.setSelectionRange(position+emoji.length,position+emoji.length);}}>{emoji}</button>{/each}</div>{/if}</div>

   </div>
   <div class="submit-actions"><span class="character-count">{body.length}/2000</span>{#if editing||parent}<button type="button" class="cancel-composer" disabled={busy} onclick={cancelComposer}>{copy($locale,"Бас тарту")}</button>{/if}<button class="button primary" disabled={busy}>{copy($locale,"Жіберу")}</button></div>
  </div>
 </form>
{/snippet}
{#snippet commentContent(item:Comment,isReply=false)}
 <header class="comment-header"><strong>{author(item)}</strong><time datetime={item.createdAt}>{new Date(item.createdAt).toLocaleDateString($locale==='ru'?'ru-RU':'kk-KZ')}</time></header>
 {#if isReply&&item.parentId!==item.rootId}<span class="reply-address">↳ {author(comments.find(comment=>comment.id===item.parentId)??comments.find(comment=>comment.id===item.rootId)??item)}</span>{/if}
 {#if item.hidden}<p class="muted">{copy($locale,isReply?"Жауап жасырылған немесе өшірілген":"Пікір жасырылған немесе өшірілген")}</p>
 {:else if item.spoiler}<details><summary>{copy($locale,"Спойлерді көрсету")}</summary><CommentBody body={item.body}/></details>
 {:else}<CommentBody body={item.body}/>{/if}
 <div class="comment-actions">
  {#if account&&!item.hidden}
   <button disabled={busy} onclick={event=>replyTo(item,event.currentTarget)}>{copy($locale,"Жауап беру")}</button>
   {#if item.userId===account}<button disabled={busy} onclick={event=>editComment(item,event.currentTarget)}>{copy($locale,"Өңдеу")}</button><button onclick={()=>remove(item.id)} disabled={busy}>{copy($locale,"Өшіру")}</button>{/if}
  {/if}
  {#if isReply||account&&!item.hidden}<button class="report-comment" onclick={()=>reporting=reporting===item.id?'':item.id}>{copy($locale,"Шағым")}</button>{/if}
 </div>
 {#if account&&(parent===item.id||editing===item.id)}{@render composer()}{/if}
 {#if reporting===item.id}<ReportForm {gateway} {titleId} {releaseId} targetType="comment" targetId={item.id}/>{/if}
{/snippet}
{#if page.url.searchParams.has('returnTo')}<a class="button subtle" href={safeQueueReturn(page.url.searchParams.get('returnTo'),'/moderation/reports')}>{text($locale,'Кезекке оралу','Вернуться в очередь')}</a>{/if}
{#if targetMissing}<p role="status">{text($locale,'Пікір қолжетімсіз немесе басқа талқылауға тиесілі.','Комментарий недоступен или относится к другому обсуждению.')}</p>{/if}
<section class="discussion" aria-label={copy($locale,"Пікірлер")}>
 <h2>{copy($locale,"Пікірлер")}</h2>
 {#if !opened}<p role="status">{copy($locale,"Жүктелуде…")}</p>{:else}
  {#if account}{#if !parent&&!editing}
   {#if !rootOpen}<button bind:this={rootTrigger} class="composer-prompt" type="button" disabled={busy} aria-expanded="false" aria-controls={rootEditorId} onclick={event=>openRoot(event.currentTarget)}><span class="prompt-icon" aria-hidden="true"><Icon name="message" size={20}/></span><span>{rootDraft.body?text($locale,'Пікір жобасын жалғастыру…','Продолжить черновик…'):text($locale,'Пікір жазу…','Написать комментарий…')}</span></button>{/if}
   <div id={rootEditorId}>{#if rootOpen}{@render composer()}{/if}</div>
  {/if}{:else}<p class="muted">{text($locale,'Пікір жазу үшін ','Чтобы написать комментарий, ')}<a class="auth-link" href={'/login?returnTo='+encodeURIComponent(releaseId?'/read/'+releaseId:titleSlug?'/titles/'+titleSlug:'/catalog')}>{copy($locale,"кіріңіз")}</a>{copy($locale," және поштаңызды растаңыз.")}</p>{/if}
  {#each comments.filter(item=>!item.parentId).sort((left,right)=>right.createdAt.localeCompare(left.createdAt)) as root}
   <article id={'comment-'+root.id} tabindex="-1" class:focused={focused===root.id}>
    <div class="comment-row">{@render avatar(root)}<div class="comment-content">{@render commentContent(root)}</div></div>
    {#if comments.some(item=>item.rootId===root.id)||root.replies>0}
     <button class="thread-toggle" type="button" disabled={busy} aria-expanded={expandedRoots.includes(root.id)} aria-controls={threadId(root.id)} onclick={()=>toggleThread(root)}><Icon name={expandedRoots.includes(root.id)?'up':'down'} size={16}/>{expandedRoots.includes(root.id)?text($locale,'Жауаптарды жию','Скрыть ответы'):text($locale,'Жауаптарды көрсету','Показать ответы')}<span class="reply-count">{replyCount(root)}</span></button>
     <div class="thread" id={threadId(root.id)} hidden={!expandedRoots.includes(root.id)}>
     {#each comments.filter(item=>item.rootId===root.id) as reply}
      <div class="reply comment-row" id={'comment-'+reply.id} tabindex="-1" class:focused={focused===reply.id}>
       {@render avatar(reply)}<div class="comment-content">{@render commentContent(reply,true)}</div>
      </div>
     {/each}
     {#if root.replies>comments.filter(item=>item.rootId===root.id).length}<button class="button subtle more-replies" disabled={busy} onclick={()=>replies(root)}>{copy($locale,"Қалған жауаптарды ашу")}</button>{/if}
    </div>{/if}
   </article>
  {:else}{#if !busy&&!loadFailed}<p class="muted">{copy($locale,"Бірінші пікірді жазыңыз.")}</p>{/if}{/each}
  {#if loadFailed}<button class="button subtle" onclick={()=>load()} disabled={busy}>{text($locale,'Қайталау','Повторить')}</button>{/if}
  {#if next}<button class="button subtle" onclick={()=>load(next)} disabled={busy}>{copy($locale,"Келесі пікірлер")}</button>{/if}
  {#if busy}<p role="status">{copy($locale,"Жүктелуде…")}</p>{/if}
 {/if}
 {#if message}<p class="form-message" role="status">{noticeText($locale,message)}</p>{/if}
 {#if showReport}<ReportForm {gateway} {titleId} {releaseId} targetType={releaseId?'release':'title'} targetId={releaseId||titleId}/>{/if}
</section>
<style>
 .discussion{padding:24px 0;max-width:800px;margin:auto;text-align:left;min-width:0}
 .composer-prompt{display:flex;align-items:center;gap:12px;width:100%;min-height:56px;padding:8px 12px;border:1px solid var(--line);border-radius:10px;background:var(--panel);color:var(--muted);font:inherit;font-size:14px;text-align:left;cursor:pointer}
 .prompt-icon{display:grid;place-items:center;flex:none;width:32px;height:32px;border-radius:50%;background:var(--raised)}
 .composer-prompt:hover{color:var(--text);border-color:var(--muted)}
 .thread-toggle{display:flex;align-items:center;gap:6px;min-height:44px;margin-left:48px;padding:6px 0;border:0;background:none;color:var(--muted);font:inherit;font-size:12px;cursor:pointer}
 .thread-toggle:hover{color:var(--text)}.reply-count{font-variant-numeric:tabular-nums}.thread[hidden]{display:none}
 .discussion h2{font-size:21px;margin:0 0 18px}
 .focused{outline:2px solid var(--accent);outline-offset:3px;scroll-margin-top:100px}
 article{border-bottom:1px solid var(--line);padding:20px 0;min-width:0}
 .comment-row{display:grid;grid-template-columns:36px minmax(0,1fr);gap:12px;align-items:start}
 .comment-content{min-width:0}
 .avatar{width:36px;height:36px;display:grid;place-items:center;overflow:hidden;border-radius:50%;background:var(--raised);color:var(--muted);font-size:14px;font-weight:600}
 .avatar img{width:100%;height:100%;object-fit:cover}
 .comment-header{display:flex;align-items:baseline;flex-wrap:wrap;gap:4px 12px;min-height:24px;line-height:1.5}
 .comment-header strong{font-size:14px;overflow-wrap:anywhere}
 time{color:var(--muted);font-size:12px;white-space:nowrap}
 p{white-space:pre-wrap;overflow-wrap:anywhere;font-size:14px}
 summary{display:flex;align-items:center;min-height:44px;cursor:pointer;overflow-wrap:anywhere}
 .comment-actions{display:flex;flex-wrap:wrap;gap:0 8px;margin-left:-8px}
 .comment-actions button,.cancel-composer{min-height:44px;max-width:100%;border:0;background:none;color:var(--muted);padding:8px;font:inherit;font-size:12px;overflow-wrap:anywhere;cursor:pointer}
 .comment-actions button:hover,.cancel-composer:hover{color:var(--text)}
 .comment-actions button:first-child{color:var(--text)}
 .comment-actions .report-comment{margin-left:auto}
 .thread{margin:8px 0 0 17px;padding-left:18px;border-left:1px solid var(--line)}
 .reply{padding-top:16px}
 .reply-address{display:block;color:var(--muted);font-size:12px;overflow-wrap:anywhere;margin:0 0 8px}
 .more-replies{margin-top:12px;white-space:normal;max-width:100%;text-align:left}
 .comment-composer{display:grid;grid-template-columns:minmax(0,1fr);gap:8px;min-width:0;margin:12px 0;padding:10px;border:1px solid var(--line);border-radius:10px;background:var(--panel)}
 .comment-composer{position:relative;padding:12px;gap:8px}
 .comment-composer:focus-within{border-color:var(--accent)}
 .composer-context{display:flex;min-width:0;gap:8px;align-items:center;border-bottom:0;padding:0 40px 0 0;min-height:24px}
 .composer-context p{flex:1;min-width:0;margin:0;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;color:var(--muted);font-size:13px}
 .composer-context strong{color:var(--text);font-size:inherit}
 .root-context{height:0;min-height:0;padding:0}.close-composer{position:absolute;right:4px;top:4px;flex:none;width:44px;height:44px;border:0;border-radius:8px;background:none;color:var(--muted);font:inherit;font-size:24px;cursor:pointer}
 .sr-only{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip-path:inset(50%);white-space:nowrap;border:0}
 label{display:grid;gap:8px;font-size:14px;min-width:0}
 .comment-composer textarea{box-sizing:border-box;width:100%;max-width:100%;min-width:0;min-height:96px;max-height:320px;resize:vertical;border:0;border-radius:0;background:transparent;padding:4px 38px 4px 2px;box-shadow:none}
 .comment-composer textarea:focus,.comment-composer textarea:focus-visible{outline:none;border:0;box-shadow:none}
 .editor-footer{border-top:1px solid var(--line);padding-top:6px;display:flex;align-items:center;flex-wrap:wrap;gap:8px;min-width:0}
 .format-actions{display:flex;align-items:center;gap:2px;flex-wrap:wrap;min-width:0;flex:1}
 .format-actions button{min-width:44px;min-height:44px;padding:8px;border:0;border-radius:6px;background:transparent;color:var(--text);font:inherit;font-size:13px;cursor:pointer}
 .format-actions button:hover,.close-composer:hover{background:var(--raised)}
 .format-actions button[aria-pressed="true"]{background:var(--raised);color:var(--accent);box-shadow:inset 0 0 0 1px var(--accent)}
 .submit-actions{display:flex;align-items:center;justify-content:flex-end;flex-wrap:wrap;gap:8px;margin-left:auto}
 .character-count{font-size:12px;color:var(--muted);font-variant-numeric:tabular-nums}
 .cancel-composer{justify-self:end;padding:0 8px}
 @keyframes composer-enter{from{opacity:.6;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}
 .comment-composer{animation:composer-enter 160ms ease-out}
 @media(prefers-reduced-motion:reduce){.comment-composer{animation:none}}
 button:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
 @media(max-width:479px){.thread-toggle{margin-left:36px}.comment-row{grid-template-columns:28px minmax(0,1fr);gap:8px}.avatar{width:28px;height:28px;font-size:12px}.thread{margin-left:13px;padding-left:10px}.comment-composer{padding:10px}.comment-content .comment-composer{margin-left:-36px}.submit-actions{flex-basis:100%}}
.emoji-picker{position:relative}.emoji-picker>button{display:grid;place-items:center;width:44px;height:44px;cursor:pointer;list-style:none;border-radius:6px}.emoji-options{position:absolute;z-index:5;bottom:50px;left:0;display:grid;grid-template-columns:repeat(4,44px);gap:2px;padding:6px;border:1px solid var(--line);border-radius:10px;background:var(--panel);box-shadow:0 8px 32px #0004}.emoji-options button{font-size:20px}.spoiler-notice{font-size:12px;color:var(--accent);margin:0;padding:4px 0}.emoji-picker>button:focus-visible{outline:2px solid var(--accent)}
</style>
