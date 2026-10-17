<script lang="ts">
 import {provideLocale} from '$lib/shared/presentation/locale';
 import {untrack} from 'svelte';
 import type {Locale} from '$lib/features/preferences/domain/locale';
 import HomeDiscovery from '$lib/features/catalog/presentation/HomeDiscovery.svelte';
 import type {Summary} from '$lib/features/catalog/application/catalog';
 let {language='kk',available=true,empty=false,shelf='updated'}:{language?:Locale;available?:boolean;empty?:boolean;shelf?:string}=$props();
 provideLocale(untrack(()=>language),()=>{});
 const title:Summary={id:'title',slug:'example',name:'Тайтл',kind:'manga',originalStatus:'ongoing',translationStatus:'ongoing',coverURL:'/api/v1/media/cover',demo:false,chapterCount:1,ratingCount:0,score:0,updatedAt:'2026-10-17T00:00:00Z'};
</script>
<HomeDiscovery catalogue={{available,page:{items:empty?[]:[title],nextCursor:null}}} updates={{available:true,page:{items:[{releaseId:'release',titleId:'title',titleName:'Тайтл',titleSlug:'example',chapterLabel:'1.5',teamName:'Команда',teamSlug:'team',demo:false,publishedAt:'2026-10-17T00:00:00Z',coverURL:'/api/v1/media/cover'}],nextCursor:null}}} picks={{available:true,page:{items:[],nextCursor:null}}} {shelf}/>
