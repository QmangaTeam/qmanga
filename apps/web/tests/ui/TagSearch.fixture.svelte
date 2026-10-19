<script lang="ts">
 import {provideLocale} from '$lib/shared/presentation/locale';
 import {untrack} from 'svelte';
 import type {Locale} from '$lib/features/preferences/domain/locale';
 import TagSearch from '$lib/features/catalog/presentation/TagSearch.svelte';
 let {language='kk',selected=[]}:{language?:Locale;selected?:string[]}=$props();
 provideLocale(untrack(()=>language),()=>{});
 const genres=[
  {id:'tag-horror',name:'Қорқыныш',nameRu:'Ужасы',slug:'horror',categorySlug:'themes'},
  {id:'tag-magic',name:'Магия',nameRu:'Магия',slug:'magic',categorySlug:'themes'},
  {id:'tag-long-hair',name:'Ұзын шаш',nameRu:'Длинные волосы',slug:'long-hair',categorySlug:'elements',parentSlug:'short-hair'},
 ];
 const categories=[{id:'cat-themes',name:'Тақырыптар',nameRu:'Темы',slug:'themes'},{id:'cat-elements',name:'Элементтер',nameRu:'Элементы',slug:'elements'}];
</script>
<TagSearch {genres} {categories} {selected} onchange={()=>{}}/>
