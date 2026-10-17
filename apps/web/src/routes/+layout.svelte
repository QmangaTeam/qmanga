<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';

	let { children } = $props();

	const navItems = [
		{ label: 'Басты бет', href: '/', icon: '🏠' },
		{ label: 'Каталог', href: '/catalog', icon: '📚' },
		{ label: 'Кітапхана', href: '/library', icon: '🔖' },
		{ label: 'Хабарламалар', href: '/notifications', icon: '🔔' },
		{ label: 'Мәзір', href: '/settings', icon: '☰' }
	];
</script>

<div class="min-h-screen bg-zinc-950 text-zinc-100 flex flex-col font-sans">
	<!-- Top Bar -->
	<header class="sticky top-0 z-40 w-full border-b border-zinc-800/80 bg-zinc-950/90 backdrop-blur">
		<div class="max-w-md md:max-w-3xl mx-auto px-4 h-14 flex items-center justify-between">
			<a href="/" class="touch-target flex items-center gap-2 font-bold text-xl tracking-tight text-amber-500">
				<span class="text-2xl">📖</span>
				<span>Qmanga</span>
			</a>
			<div class="flex items-center gap-1">
				<a
					href="/catalog"
					aria-label="Іздеу"
					class="touch-target rounded-full hover:bg-zinc-800 text-zinc-400 hover:text-zinc-100 text-lg"
				>
					🔍
				</a>
			</div>
		</div>
	</header>

	<!-- Main Content Area -->
	<main class="flex-1 w-full max-w-md md:max-w-3xl mx-auto px-4 py-4 pb-24">
		{@render children()}
	</main>

	<!-- Mobile Bottom Navigation (Senkuro-style 5 buttons) -->
	<nav class="fixed bottom-0 left-0 right-0 z-40 border-t border-zinc-800/80 bg-zinc-950/95 backdrop-blur pb-safe">
		<div class="max-w-md md:max-w-3xl mx-auto flex items-center justify-around h-16 px-1">
			{#each navItems as item}
				{@const active = $page.url.pathname === item.href}
				<a
					href={item.href}
					class="touch-target flex flex-col items-center justify-center flex-1 py-1 px-1 rounded-lg transition-colors {active ? 'text-amber-500 font-medium' : 'text-zinc-400 hover:text-zinc-200'}"
				>
					<span class="text-lg leading-none">{item.icon}</span>
					<span class="text-[11px] leading-tight mt-1 truncate max-w-[64px]">{item.label}</span>
				</a>
			{/each}
		</div>
	</nav>
</div>
