<script lang="ts">
 import type {Snippet} from 'svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from './Icon.svelte';
 let {title,description,children,footer}:{title:string;description?:string;children:Snippet;footer?:Snippet}=$props();
 const {locale}=useLocale();
</script>
<div class="auth-page"><a class="auth-back" href="/"><Icon name="left" size={16}/>{text($locale,'Басты бетке','На главную')}</a><section class="auth-layout"><a class="auth-brand" href="/" aria-label="Qmanga"><img src="/favicon.png" alt="" width="32" height="32"/></a><header><h1>{title}</h1>{#if description}<p>{description}</p>{/if}</header>{@render children()}{#if footer}<div class="auth-footer">{@render footer()}</div>{/if}</section></div>
<style>
 .auth-page{max-width:440px;margin:0 auto 32px}.auth-back{display:inline-flex;align-items:center;gap:8px;min-height:44px;margin-bottom:12px;color:var(--muted);font-size:12px;text-decoration:none}.auth-layout{border:1px solid var(--line);background:var(--panel);border-radius:12px;padding:24px 20px}.auth-brand{display:grid;place-items:center;width:44px;min-height:44px;border:1px solid var(--line);border-radius:9px;margin-bottom:20px}.auth-brand img{display:block}header{margin-bottom:24px}h1{font-size:25px;line-height:1.2;margin:0;letter-spacing:-.5px}header p{font-size:13px;line-height:1.6;color:var(--muted);margin:10px 0 0}.auth-footer{border-top:1px solid var(--line);margin-top:24px;padding-top:16px;font-size:12px;line-height:1.7;color:var(--muted)}@media(min-width:768px){.auth-page{margin-top:20px}.auth-layout{padding:30px}}
</style>
