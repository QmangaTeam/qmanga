<script lang="ts">
 import Notifications from '$lib/features/notifications/presentation/Notifications.svelte';
 import {createNotifications} from '$lib/composition/browser';
 const gateway=createNotifications();
</script>
<Notifications {gateway}/>
