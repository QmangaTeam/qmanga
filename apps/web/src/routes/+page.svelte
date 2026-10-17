<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Qmanga — Басты бет</title>
</svelte:head>

<div class="space-y-6">
	<!-- Hero Section -->
	<section class="rounded-2xl bg-gradient-to-br from-zinc-900 to-zinc-900/60 p-4 border border-zinc-800">
		<div class="flex items-start justify-between gap-2">
			<div>
				<span class="inline-block px-2.5 py-0.5 text-xs font-semibold rounded-full bg-amber-500/20 text-amber-400 border border-amber-500/30">
					Қазақ тіліндегі манга
				</span>
				<h1 class="text-xl xs:text-2xl font-extrabold mt-2 tracking-tight text-white">
					Qmanga платформасына қош келдіңіз!
				</h1>
				<p class="text-zinc-400 text-sm mt-1 leading-relaxed">
					Манга, манхва және маньхуаны ана тілімізде жылдам әрі ыңғайлы оқыңыз.
				</p>
			</div>
		</div>

		<div class="mt-4 flex flex-wrap gap-2.5">
			<a
				href="/catalog"
				class="touch-target px-5 rounded-xl bg-amber-500 hover:bg-amber-400 text-zinc-950 font-bold text-sm shadow-sm transition active:scale-[0.98]"
			>
				Каталогты қарау
			</a>
			<a
				href="/library"
				class="touch-target px-4 rounded-xl bg-zinc-800 hover:bg-zinc-700 text-zinc-200 font-medium text-sm transition active:scale-[0.98]"
			>
				Менің кітапханам
			</a>
		</div>
	</section>

	<!-- Backend & Infrastructure Health Status Card -->
	<section class="rounded-xl bg-zinc-900/80 border border-zinc-800 p-4">
		<div class="flex items-center justify-between pb-3 border-b border-zinc-800/80">
			<h2 class="text-sm font-bold uppercase tracking-wider text-zinc-400 flex items-center gap-2">
				<span class="relative flex h-2.5 w-2.5">
					{#if data.apiConnected && data.health?.status === 'ok'}
						<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
						<span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
					{:else}
						<span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-amber-500"></span>
					{/if}
				</span>
				Жүйелік байланыс күйі
			</h2>
			<span class="text-xs px-2 py-0.5 rounded bg-zinc-800 text-zinc-300 font-mono">
				v{data.health?.version || '0.1.0'}
			</span>
		</div>

		<div class="grid grid-cols-1 xs:grid-cols-2 gap-3 mt-3 text-sm">
			<div class="bg-zinc-950/60 p-3 rounded-lg border border-zinc-800/60">
				<div class="text-xs text-zinc-400">Go API қызметі</div>
				<div class="font-semibold mt-1 flex items-center gap-1.5">
					{#if data.apiConnected}
						<span class="text-emerald-400">✓ Қосылған</span>
					{:else}
						<span class="text-rose-400">✗ {data.error || 'Байланыс жоқ'}</span>
					{/if}
				</div>
			</div>

			<div class="bg-zinc-950/60 p-3 rounded-lg border border-zinc-800/60">
				<div class="text-xs text-zinc-400">PostgreSQL дерекқоры</div>
				<div class="font-semibold mt-1 flex items-center gap-1.5">
					{#if data.health?.postgres === 'ok'}
						<span class="text-emerald-400">✓ Жұмыс істеп тұр</span>
					{:else}
						<span class="text-amber-400">⟳ {data.health?.postgres || 'Күтуде...'}</span>
					{/if}
				</div>
			</div>

			<div class="bg-zinc-950/60 p-3 rounded-lg border border-zinc-800/60">
				<div class="text-xs text-zinc-400">MinIO S3 сақтау орны</div>
				<div class="font-semibold mt-1 flex items-center gap-1.5">
					{#if data.health?.s3 === 'ok'}
						<span class="text-emerald-400">✓ Жұмыс істеп тұр</span>
					{:else}
						<span class="text-amber-400">⟳ {data.health?.s3 || 'Күтуде...'}</span>
					{/if}
				</div>
			</div>

			<div class="bg-zinc-950/60 p-3 rounded-lg border border-zinc-800/60">
				<div class="text-xs text-zinc-400">Web SSR (Bun)</div>
				<div class="font-semibold mt-1 text-emerald-400 flex items-center gap-1.5">
					✓ Қалыпты (Alpine)
				</div>
			</div>
		</div>
	</section>

	<!-- Mobile Demo Manga Cards (2 columns on mobile phone layout) -->
	<section class="space-y-3">
		<div class="flex items-center justify-between">
			<h2 class="text-base font-bold text-zinc-100">Жаңа шығарылымдар</h2>
			<a href="/catalog" class="text-xs text-amber-500 hover:underline touch-target">
				Барлығын көру →
			</a>
		</div>

		<div class="grid grid-cols-2 gap-3">
			<div class="group rounded-xl bg-zinc-900 border border-zinc-800/80 overflow-hidden flex flex-col">
				<div class="aspect-[3/4] bg-zinc-800 flex items-center justify-center text-4xl select-none group-hover:scale-105 transition-transform duration-300">
					🥋
				</div>
				<div class="p-2.5 flex-1 flex flex-col justify-between">
					<div>
						<span class="text-[10px] uppercase font-bold text-amber-400">Манхва</span>
						<h3 class="text-xs xs:text-sm font-semibold text-zinc-100 line-clamp-1 mt-0.5">
							Жалғыз деңгей көтеру
						</h3>
					</div>
					<div class="mt-2 text-[11px] text-zinc-400 flex items-center justify-between">
						<span>1-тарау</span>
						<span class="text-zinc-400">Q-Team</span>
					</div>
				</div>
			</div>

			<div class="group rounded-xl bg-zinc-900 border border-zinc-800/80 overflow-hidden flex flex-col">
				<div class="aspect-[3/4] bg-zinc-800 flex items-center justify-center text-4xl select-none group-hover:scale-105 transition-transform duration-300">
					⚔️
				</div>
				<div class="p-2.5 flex-1 flex flex-col justify-between">
					<div>
						<span class="text-[10px] uppercase font-bold text-amber-400">Манга</span>
						<h3 class="text-xs xs:text-sm font-semibold text-zinc-100 line-clamp-1 mt-0.5">
							Бүркітші бала
						</h3>
					</div>
					<div class="mt-2 text-[11px] text-zinc-400 flex items-center justify-between">
						<span>1-тарау</span>
						<span class="text-zinc-400">Дала тобы</span>
					</div>
				</div>
			</div>
		</div>
	</section>
</div>
