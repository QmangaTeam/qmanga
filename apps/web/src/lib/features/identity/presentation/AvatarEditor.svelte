<script lang="ts">
 import {onDestroy,untrack} from 'svelte';
 import {avatarSourceValid,type AvatarGateway} from '../application/avatars';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureText} from '$lib/shared/presentation/messages';
 let {gateway,kind,ownerId,name,actorId}:{gateway:AvatarGateway;kind:'user'|'team';ownerId:string;name:string;actorId:string}=$props();
 const {locale}=useLocale();let busy=$state(false),message=$state(''),failed=$state(false),version=$state(0);
 let pending=$state<File|null>(null),preview=$state('');
 function cancel(){if(preview)URL.revokeObjectURL(preview);preview='';pending=null;}
 onDestroy(cancel);
 $effect(()=>{actorId;ownerId;untrack(cancel);});
 const initials=$derived([...name.trim()].slice(0,2).join('').toUpperCase()||'Q');
 function upload(event:Event){const input=event.currentTarget as HTMLInputElement;const file=input.files?.[0];input.value='';if(!file||busy)return;cancel();message='';if(!avatarSourceValid(file)){message=text($locale,'JPEG/PNG/WebP, 5 МБ дейін.','JPEG/PNG/WebP, до 5 МБ.');return;}pending=file;preview=URL.createObjectURL(file);}
 async function confirm(){if(!pending||busy)return;const file=pending,actor=actorId,owner=ownerId;busy=true;try{await gateway.upload(kind,owner,file,actor);if(actor!==actorId||owner!==ownerId)return;failed=false;version=Date.now();message=text($locale,'Аватар сақталды.','Аватар сохранён.');cancel();}catch(error){if(actor===actorId&&owner===ownerId)message=failureText($locale,error);}finally{busy=false;}}

</script>
<div class="avatar-editor"><span class="avatar" aria-hidden="true">{#if !failed}<img src={'/media/avatars/'+kind+'/'+encodeURIComponent(ownerId)+'?v='+version} alt="" onerror={()=>failed=true}/>{:else}{initials}{/if}</span><label>{text($locale,'Аватарды өзгерту','Изменить аватар')}<input type="file" accept="image/jpeg,image/png,image/webp" disabled={busy} onchange={upload}/><span class="muted">{text($locale,'JPEG/PNG/WebP, 5 МБ дейін. Ортадан шаршы қиылады.','JPEG/PNG/WebP, до 5 МБ. Обрезается квадратом по центру.')}</span></label></div>
{#if pending}<section class="crop-preview" aria-label={text($locale,'Аватарды алдын ала қарау','Предпросмотр аватара')}><p>{text($locale,'Сақталатын орталық шаршы және дөңгелек көрініс.','Центральная квадратная обрезка и круглый вид после сохранения.')}</p><div><img src={preview} alt={text($locale,'Шаршы қиынды','Квадратная обрезка')}/><img class="round" src={preview} alt={text($locale,'Дөңгелек көрініс','Круглый вид')}/></div><button class="button primary" disabled={busy} onclick={confirm}>{text($locale,'Сақтау','Сохранить')}</button><button class="button" disabled={busy} onclick={cancel}>{text($locale,'Бас тарту','Отмена')}</button></section>{/if}
{#if message}<p role="status">{message}</p>{/if}
<style>.crop-preview{padding:12px;border:1px solid var(--line)}.crop-preview>div{display:flex;gap:12px;margin:12px 0}.crop-preview img{width:112px;height:112px;object-fit:cover;object-position:center}.crop-preview .round{border-radius:50%}.crop-preview button{margin:4px}.avatar-editor{display:flex;align-items:center;gap:16px;margin:20px 0;min-width:0}.avatar{display:grid;place-items:center;flex-shrink:0;width:64px;height:64px;border-radius:50%;background:var(--panel);border:1px solid var(--line);font-weight:650}.avatar img{width:64px;height:64px;object-fit:cover;border-radius:50%}label{display:grid;gap:8px;min-width:0}input{max-width:100%;min-height:44px}.muted{font-size:12px;color:var(--muted)}</style>
