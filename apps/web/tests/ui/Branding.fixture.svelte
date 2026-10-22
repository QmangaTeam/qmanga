<script lang="ts">
 import AuthLayout from '$lib/shared/presentation/ui/AuthLayout.svelte';
 import MobileNavigation from '$lib/shared/presentation/ui/MobileNavigation.svelte';
 let {currentPath='/'}:{currentPath?:string}=$props();
 const items=[{label:'Басты бет',href:'/',icon:'home' as const},{label:'Каталог',href:'/catalog',icon:'catalog' as const},{label:'Кітапхана',shortLabel:'Сөре',href:'/library',icon:'bookmark' as const},{label:'Хабарлар',href:'/notifications',icon:'bell' as const},{label:'Мәзір',href:'/menu',icon:'menu' as const}];
</script>
<AuthLayout title="Кіру">{#snippet children()}<form action="/login?returnTo=%2Flibrary" method="POST"><input name="email"/></form>{/snippet}</AuthLayout>
<MobileNavigation {items} {currentPath} label="Mobile" unreadCount={12} unreadBadge="12"/>
