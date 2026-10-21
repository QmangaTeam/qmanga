<script lang="ts">
 import {untrack} from 'svelte';
 import {provideLocale} from '$lib/shared/presentation/locale';
 import Title from '$lib/features/catalog/presentation/Title.svelte';
 import type {TitleDetail} from '$lib/features/catalog/application/catalog';
 let {bannerURL='',coverURL='/api/v1/media/cover',language='ru',tab='chapters',ageRating=0,genreCount=0,editionCount=1}:{genreCount?:number;editionCount?:number;bannerURL?:string;coverURL?:string;language?:'kk'|'ru';tab?:string;ageRating?:0|12|16|18}=$props();
 provideLocale(untrack(()=>language),()=>{});
 const title=$derived<TitleDetail>({bannerURL,ageRating,id:'title',slug:'example',name:'Ұзақ атауы бар тайтл',originalName:'Original title',aliases:['Another title'],kind:'manhwa',originalStatus:'ongoing',translationStatus:'ongoing',description:'Описание произведения',direction:'ltr',demo:true,coverURL,score:8.5,ratingCount:2,genres:Array.from({length:genreCount},(_,i)=>({id:`genre-${i}`,slug:`genre-${i}`,name:`Жанр ${i+1}`})),projects:[{id:'project',titleId:'title',teamId:'team',titleName:'Ұзақ атауы бар тайтл',titleSlug:'example',teamName:'Тест командасы',teamSlug:'test-team',status:'active'}],chapterCount:185,commentCount:12,readerCount:12,publication:{year:2023,publishers:['Publisher'],source:'Новелла',country:'Корея',formats:['В цвете','Вебтун'],links:[{label:'Официальный сайт',url:'https://example.com/title'}]}});
 const chapter=$derived({id:'chapter',label:'185',volume:'1',name:'',order:185,releases:Array.from({length:editionCount},(_,i)=>({id:i?`release-${i}`:'release',projectId:'project',teamId:`team-${i}`,teamName:`Команда ${i+1}${i%3===0?' — ұзақ аударма командасының атауы'.repeat(3):''}`,teamSlug:`team-${i}`,publishedAt:'2026-10-18T00:00:00Z'}))});
</script>
<Title {title} chapters={{items:[chapter],nextCursor:null}} nextURL={null} firstChapter={chapter} {tab}>
 {#snippet library()}<button>Добавить в список</button>{/snippet}
 {#snippet discussion()}<p>Тестовые комментарии</p>{/snippet}
 {#snippet report()}<button>Нашёл ошибку</button>{/snippet}
</Title>
