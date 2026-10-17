<script lang="ts">
 import Moderation from '$lib/features/publishing/presentation/Moderation.svelte';
 import { createPublishing } from '$lib/composition/browser';
 import type { PageData } from './$types';
 const gateway = createPublishing(); let { data }: { data: PageData } = $props();
</script>
<a class="auth-link" href="/moderation/reports">Шағымдарды қарау</a>
<Moderation user={data.user} {gateway}/>
