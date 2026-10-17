<script lang="ts">
 import CoverCard from '$lib/features/catalog/presentation/CoverCard.svelte';
 import ChapterRow from '$lib/features/catalog/presentation/ChapterRow.svelte';
 import LibraryRow from '$lib/features/reading/presentation/LibraryRow.svelte';
 import MobileNavigation from '$lib/shared/presentation/ui/MobileNavigation.svelte';
 let {releaseId=''}:{releaseId?:string}=$props();
</script>
<CoverCard href="/titles/example" name="Үлгі" coverURL="/api/v1/media/example"/>
<ChapterRow title="1-тарау" releases={[{id:'release',teamName:'Команда',teamSlug:'team'}]} readLabel="Оқу"/>
<div data-library><LibraryRow name="Үлгі" href="/titles/example" {releaseId} continueLabel="Жалғастыру"/></div>
<MobileNavigation items={[{label:'Каталог',href:'/catalog',icon:'catalog'}]} currentPath="/catalog" label="Навигация"/>
