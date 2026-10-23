<script lang="ts">
 import {provideLocale} from '$lib/shared/presentation/locale';
 import {untrack} from 'svelte';
 import type {Locale} from '$lib/features/preferences/domain/locale';
 import HomeDiscovery from '$lib/features/catalog/presentation/HomeDiscovery.svelte';
 import type {Summary} from '$lib/features/catalog/application/catalog';
 let {language='kk',available=true,empty=false,topKind='manhwa',topFilled=false,hasContinuation=false,topRequested=false,size=1,distinctTop=false,updateCount=1}:{language?:Locale;available?:boolean;empty?:boolean;topKind?:string;topFilled?:boolean;hasContinuation?:boolean;topRequested?:boolean;size?:number;distinctTop?:boolean;updateCount?:number}=$props();
 provideLocale(untrack(()=>language),()=>{});
 const title:Summary={id:'title',slug:'example',name:'Тайтл',kind:'manga',originalStatus:'ongoing',translationStatus:'ongoing',coverURL:'/api/v1/media/cover',demo:false,chapterCount:1,ratingCount:0,score:0,updatedAt:'2026-10-17T00:00:00Z'};
</script>
<HomeDiscovery mostRead={{available,page:{items:empty?[]:Array.from({length:size},(_,index)=>index?{...title,id:'title-'+index,slug:'example-'+index}:title),nextCursor:null}}} updates={{available:true,page:{items:Array.from({length:updateCount},(_,index)=>({releaseId:index?'release-'+index:'release',titleId:'title',titleName:'Тайтл',titleSlug:'example',chapterLabel:'1.5',teamName:'Команда',teamSlug:'team',demo:false,publishedAt:'2026-10-17T00:00:00Z',coverURL:'/api/v1/media/cover'})),nextCursor:null}}} top={{available:true,page:{items:topFilled?[distinctTop?{...title,id:'different',slug:'different'}:title]:[],nextCursor:null}}} {topKind} {hasContinuation} {topRequested}>
 {#snippet continuation()}<a href="/library">Кітапхана</a>{/snippet}
</HomeDiscovery>
