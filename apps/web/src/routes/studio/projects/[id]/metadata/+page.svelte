<script lang="ts">
 import {projectListReturn} from '$lib/features/publishing/domain/workspace-context';
 import {page} from '$app/stores';
 import ProjectNavigation from '$lib/features/publishing/presentation/ProjectNavigation.svelte';
 import TitleChangeForm from '$lib/features/catalog/presentation/TitleChangeForm.svelte';
 import {createTitleChanges,createMetadataImports} from '$lib/composition/browser';
 const gateway=createTitleChanges();
</script>
<ProjectNavigation projectId={$page.params.id||''} current="metadata" returnTo={$page.url.searchParams.get('projectsReturn')||''}/>
{#key $page.params.id}<TitleChangeForm imports={createMetadataImports()} {gateway} projectId={$page.params.id||''} projectReturn={projectListReturn($page.url.searchParams.get('projectsReturn'))} coverPageId={$page.url.searchParams.get('coverPage')||''}/>{/key}
