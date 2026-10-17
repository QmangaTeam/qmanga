<script lang="ts">
 import type { CatalogFilter, Genre, ResultPage, Summary } from '../application/catalog';
 import { kindNames, originalStatusNames, translationStatusNames } from '../domain/catalog';
 let { view }: { view: { available: boolean; message: string; filter: CatalogFilter; genres: Genre[]; page: ResultPage<Summary>; nextURL: string | null } } = $props();
</script>
<svelte:head><title>Каталог — Qmanga</title><meta name="description" content="Қазақ тіліндегі манга, манхва және маньхуа каталогы."/></svelte:head>
<h1>Каталог</h1>
<form action="/catalog" method="GET" class="catalogue-search">
 <label for="catalogue-query">Тайтл немесе басқа атауы</label>
 <div class="search-row"><input id="catalogue-query" type="search" name="q" value={view.filter.query} maxlength="80" placeholder="Атауын іздеу"/><button type="submit" class="button primary">Іздеу</button></div>
 <details><summary>Сүзгілер мен реттеу</summary><div class="filter-fields">
  <label>Түрі<select name="kind" value={view.filter.kind}><option value="">Барлығы</option>{#each Object.entries(kindNames) as [value, name]}<option {value}>{name}</option>{/each}</select></label>
  <label>Жанр<select name="genre" value={view.filter.genre}><option value="">Барлығы</option>{#each view.genres as genre}<option value={genre.slug}>{genre.name}</option>{/each}</select></label>
  <label>Түпнұсқа күйі<select name="original_status" value={view.filter.originalStatus}><option value="">Барлығы</option>{#each Object.entries(originalStatusNames) as [value, name]}<option {value}>{name}</option>{/each}</select></label>
  <label>Аударма күйі<select name="translation_status" value={view.filter.translationStatus}><option value="">Барлығы</option>{#each Object.entries(translationStatusNames) as [value, name]}<option {value}>{name}</option>{/each}</select></label>
  <label>Реттеу<select name="sort" value={view.filter.sort}><option value="updated">Соңғы жаңартулар</option><option value="name">Атауы бойынша</option></select></label>
  <button type="submit" class="button primary">Қолдану</button><a class="button subtle" href="/catalog">Тазарту</a>
 </div></details>
</form>
{#if !view.available}<div class="form-message" role="alert"><p>{view.message}</p><a class="auth-link" href="/catalog">Қайта жүктеу</a></div>
{:else if view.page.items.length === 0}<div class="empty-state"><h2>Тайтл табылмады</h2><p>Басқа атаумен іздеп көріңіз немесе сүзгілерді тазалаңыз. Жарияланған аудармалар осында көрінеді.</p></div>
{:else}<div class="catalogue-grid">{#each view.page.items as title (title.id)}<a class="catalogue-card" href={'/titles/' + encodeURIComponent(title.slug)}>
 <div class="cover">{#if title.coverURL}<img src={title.coverURL} alt="" width="280" height="400" loading="lazy" decoding="async"/>{:else}<span>Мұқаба жоқ</span>{/if}</div>
 <h2>{title.name}</h2><p>{kindNames[title.kind]}{#if title.demo} / Үлгі{/if}</p><p>{title.chapterCount} тарау{#if title.ratingCount > 0} / {title.score.toFixed(1)} ★{/if}</p>
</a>{/each}</div>{/if}
{#if view.nextURL}<a class="button subtle" href={view.nextURL}>Келесі бет</a>{/if}
<style>
 .catalogue-search { margin:24px 0; display:grid; gap:10px; } .search-row { display:flex; gap:10px; } .search-row input { flex:1; }
 summary { min-height:44px; padding:10px 0; cursor:pointer; } .filter-fields { display:grid; gap:16px; margin-top:12px; } .filter-fields label { display:grid; gap:6px; }
 .catalogue-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:22px 14px; margin:24px 0; }
 .catalogue-card { min-width:0; } .cover { aspect-ratio:7/10; background:var(--panel); display:grid; place-items:center; color:var(--muted); font-size:13px; overflow:hidden; border-radius:8px; } .cover img { width:100%; height:100%; object-fit:cover; }
 .catalogue-card h2 { font-size:15px; line-height:1.4; margin-top:8px; overflow-wrap:anywhere; } .catalogue-card p { font-size:13px; color:var(--muted); margin-top:4px; }
 @media(min-width:400px) { .catalogue-grid { grid-template-columns:repeat(3,minmax(0,1fr)); } } @media(min-width:768px) { .filter-fields { grid-template-columns:repeat(3,minmax(0,1fr)); } .catalogue-grid { grid-template-columns:repeat(5,minmax(0,1fr)); } }
</style>
