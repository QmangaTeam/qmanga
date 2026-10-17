<script lang="ts">
 import Badge from '$lib/shared/presentation/ui/Badge.svelte';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import SettingsRow from '$lib/shared/presentation/ui/SettingsRow.svelte';
 import Pager from '$lib/shared/presentation/ui/Pager.svelte';
</script>
<Badge label="Дайын"/>
<SectionHeading level={1} title="Тақырып"/>
<SectionHeading level={2} title="Тараулар"/>
<SectionHeading level={3} title="Команда"/>
<div data-case="empty"><StateMessage kind="empty" title="Әзірге бос"/></div>
<div data-case="error"><StateMessage kind="error" title="Қате"/></div>
<StateMessage kind="loading" title="Күте тұр"/>
<SettingsRow controlId="quality" label="Сапа"><select id="quality"><option>Авто</option></select></SettingsRow>
<Pager label="Тараулар" previous={{label:'Артқа',action:()=>{},busy:true}} next={{label:'Келесі',href:'/catalog?cursor=test'}}/>
