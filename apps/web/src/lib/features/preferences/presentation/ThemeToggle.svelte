<script lang="ts">
 import {onMount} from 'svelte';
 import type {ThemeStore} from '../application/theme';
 import {changeTheme,initializeTheme} from '../application/theme';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {preferences}:{preferences:ThemeStore}=$props();
 const {locale}=useLocale();
 let dark=$state(true);
 function update(){dark=document.documentElement.dataset.theme==='dark'||(!document.documentElement.dataset.theme&&matchMedia('(prefers-color-scheme:dark)').matches);}
 onMount(()=>{initializeTheme(preferences);update();const observer=new MutationObserver(update);observer.observe(document.documentElement,{attributes:true,attributeFilter:['data-theme']});return()=>observer.disconnect();});
</script>
<button type="button" aria-label={dark?text($locale,'Ашық тақырыпқа ауысу','Включить светлую тему'):text($locale,'Қараңғы тақырыпқа ауысу','Включить тёмную тему')} title={text($locale,'Тақырыпты ауыстыру','Переключить тему')} onclick={()=>{changeTheme(preferences,dark?'light':'dark');update();}}><Icon name={dark?'sun':'moon'}/></button>
<style>button{display:grid;place-items:center;min-height:44px;min-width:44px;border:0;border-radius:8px;background:none;color:var(--muted);cursor:pointer}button:hover{background:var(--raised);color:var(--accent)}button:focus-visible{outline:2px solid var(--accent);outline-offset:-2px}</style>
