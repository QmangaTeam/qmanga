<script lang="ts">
 import {onMount} from 'svelte';
 import type {Downloads,Download} from '../application/downloads';
 let {downloads}:{downloads:Downloads}=$props();
 let entries=$state<Download[]>([]),message=$state(''),busy=$state(true);
 async function load(){busy=true;try{entries=await downloads.list();}catch{message='Сақталған деректер қолжетімсіз';}finally{busy=false;}}
 onMount(()=>{void load();});
 async function verify(item:Download){try{await downloads.reconcile(item);await load();}catch{message='Көшірмені тексеру мүмкін болмады';}}
 async function remove(id:string){try{await downloads.remove(id);await load();}catch{message='Көшірмені өшіру мүмкін болмады';}}
</script>
<svelte:head><title>Сақталған тараулар — Qmanga</title></svelte:head>
<h1>Сақталған тараулар</h1><p class="muted">Осы браузердегі көшірмелер. Браузер деректерді тазаласа, қайта жүктеу қажет.</p>
{#if message}<p role="status" class="form-message">{message}</p>{/if}
{#if busy}<p role="status">Жүктелуде…</p>{:else if !entries.length}<div class="empty-state"><h2>Сақталған тарау жоқ</h2><p>Читалкадағы «Баптаулар» арқылы тарауды құрылғыға сақтаңыз.</p><a href="/catalog" class="button">Каталогқа өту</a></div>{/if}
{#each entries as item}<article><h2>{item.manifest.titleName}</h2><p>{item.manifest.chapterLabel}-тарау · {item.manifest.team.name}</p><p class="muted">{item.status==='ready'?'Толық сақталған':item.status==='partial'?'Толық емес':'Сақтау қатесі'} · {item.tier} · {(item.bytes/1024/1024).toFixed(1)} / {(item.total/1024/1024).toFixed(1)} МБ</p><div>{#if item.status==='ready'}<a class="button primary" href={`/offline-reader?release=${item.manifest.releaseId}`} data-sveltekit-reload>Желісіз оқу</a>{:else}<a class="button" href={`/read/${item.manifest.releaseId}`}>Жүктеуді жалғастыру</a>{/if}<button class="button subtle" onclick={()=>verify(item)}>Тексеру</button><button class="button subtle" onclick={()=>remove(item.id)}>Көшірмені өшіру</button></div></article>{/each}
<style>h1{font-size:26px;margin-bottom:8px}article{padding:24px 0;border-bottom:1px solid var(--line)}article h2{font-size:18px}article p{font-size:14px;margin:8px 0}article>div{display:flex;flex-wrap:wrap;gap:8px;margin-top:16px}.empty-state{margin:24px 0}.empty-state h2{font-size:18px}.empty-state a{margin-top:20px}</style>
