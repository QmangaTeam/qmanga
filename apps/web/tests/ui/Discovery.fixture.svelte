<script lang="ts">
 import {provideLocale} from '$lib/shared/presentation/locale';
 import {untrack} from 'svelte';
 import type {Locale} from '$lib/features/preferences/domain/locale';
 import CoverCard from '$lib/features/catalog/presentation/CoverCard.svelte';
 import Updates from '$lib/features/catalog/presentation/Updates.svelte';
 import AppHeader from '$lib/shared/presentation/ui/AppHeader.svelte';
 let {language='kk',ratingCount=0,ageRating=0}:{language?:Locale;ratingCount?:number;ageRating?:0|12|16|18}=$props();
 provideLocale(untrack(()=>language),()=>{});
</script>
<AppHeader items={[{label:'Каталог',href:'/catalog',icon:'catalog'},{label:'Кітапхана',href:'/library',icon:'bookmark'},{label:'Хабарлар',href:'/notifications',icon:'bell'}]} currentPath="/" accountHref="/menu" accountLabel="Аккаунт" homeLabel="Qmanga" navigationLabel="Навигация"/>
<CoverCard {ageRating} href="/titles/example" name="Тайтл" coverURL="/api/v1/media/cover" metadata="Манга" score={8.25} {ratingCount}/>
<Updates items={[{releaseId:'release',titleId:'title',titleName:'Тайтл',titleSlug:'example',chapterLabel:'1.5',teamName:'Команда',teamSlug:'team',demo:false,publishedAt:'2026-10-17T00:00:00Z',coverURL:'/api/v1/media/cover'}]}/>
