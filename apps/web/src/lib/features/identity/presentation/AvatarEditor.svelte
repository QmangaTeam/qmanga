<script lang="ts">
 import {avatarSourceValid,type AvatarGateway} from '../application/avatars';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureText} from '$lib/shared/presentation/messages';
 let {gateway,kind,ownerId,name,actorId}:{gateway:AvatarGateway;kind:'user'|'team';ownerId:string;name:string;actorId:string}=$props();
 const {locale}=useLocale();let busy=$state(false),message=$state(''),failed=$state(false),version=$state(0);
 const initials=$derived([...name.trim()].slice(0,2).join('').toUpperCase()||'Q');
 async function upload(event:Event){const input=event.currentTarget as HTMLInputElement;const file=input.files?.[0];if(!file||busy)return;message='';if(!avatarSourceValid(file)){message=text($locale,'JPEG/PNG/WebP, 5 МБ дейін.','JPEG/PNG/WebP, до 5 МБ.');input.value='';return;}busy=true;try{await gateway.upload(kind,ownerId,file,actorId);failed=false;version=Date.now();message=text($locale,'Аватар сақталды.','Аватар сохранён.');}catch(error){message=failureText($locale,error);}finally{busy=false;input.value='';}}
</script>
<div class="avatar-editor"><span class="avatar" aria-hidden="true">{#if !failed}<img src={'/media/avatars/'+kind+'/'+encodeURIComponent(ownerId)+'?v='+version} alt="" onerror={()=>failed=true}/>{:else}{initials}{/if}</span><label>{text($locale,'Аватарды өзгерту','Изменить аватар')}<input type="file" accept="image/jpeg,image/png,image/webp" disabled={busy} onchange={upload}/><span class="muted">{text($locale,'JPEG/PNG/WebP, 5 МБ дейін. Ортадан шаршы қиылады.','JPEG/PNG/WebP, до 5 МБ. Обрезается квадратом по центру.')}</span></label></div>
{#if message}<p role="status">{message}</p>{/if}
<style>.avatar-editor{display:flex;align-items:center;gap:16px;margin:20px 0;min-width:0}.avatar{display:grid;place-items:center;flex-shrink:0;width:64px;height:64px;border-radius:50%;background:var(--panel);border:1px solid var(--line);font-weight:650}.avatar img{width:64px;height:64px;object-fit:cover;border-radius:50%}label{display:grid;gap:8px;min-width:0}input{max-width:100%;min-height:44px}.muted{font-size:12px;color:var(--muted)}</style>
