# Senkuro layout adaptation

Owner direction, 18 October 2026: copy Senkuro's placement and interaction patterns, retain Qmanga's palette, use less text and more clickable artwork and recognizable action controls. This replaces the rejected text-first home composition, not the backend workflow contract.

## Visual contract

### Home section order — owner's latest correction

Owner-approved refinement,23 October2026 (local-only TASK-127): one combined ranking shelf starts the homepage, then Continue reading, then bounded Latest updates. Most read and Manga/Manhwa/Manhua are real query-based switches replacing that same shelf, with one active selection. No duplicate lower type shelf. Mobile uses a2×2 switch grid; desktop one row. Empty/error states belong to the selected ranking. Only the first three selected-shelf covers are eager. This replaces the earlier returning-reader-first arrangement at the owner's explicit request.

Most read is an all-time ranking of distinct non-blocked accounts with persisted reading progress. Anonymous/offline-only local progress is not silently counted as server statistics. Type tops rank published titles by actual non-blocked-user average rating, then vote count and stable title ID. Unrated titles sort after rated titles; no score badge appears without votes. Rankings exclude unpublished announcements, and each response is bounded to16 titles. The backend read model reuses catalogue summaries and Go media URLs; no new tracking events, image provider or database table was introduced.

Only the selected type-top shelf is requested; tab links preserve the selected kind in the URL and target the lower section. Only the first three most-read covers have eager/high priority. Type-top covers below the fold are lazy. Continuation has a permanent section heading and separate loading, empty and failed-network states instead of disappearing without explanation.

Tests exercise real PostgreSQL counts/order, exclusion of blocked readers and ratings, private announcements, kind filtering and limits; frontend tests assert section order, actual type endpoint, working type links and below-fold image priority. Local checks/build/deployment are performed serially for this pass to avoid running simultaneous full image builds after the owner's memory report.

Validated on the running stack: all four headings in order in Russian and Kazakh, type switching with two real manga covers and an honest empty manhua category, lazy type-top images, zero horizontal overflow at320px, and home actions at least44px tall. Desktop was checked at1280px; Kazakh mobile at360px. The public ranking API returns three overall titles, one manhwa, two manga and zero manhua for this local dataset; invalid kinds receive422 and limits above16 receive400.

Final full Go suite passes, including real PostgreSQL ranking tests. Serial Alpine API/web builds pass41 frontend application/infrastructure tests and23 UI contracts, architecture/contrast gates, application and preview checks with zero errors/warnings. The obsolete preview selector warning was removed and rechecked, without leaving the preview server running. API/web containers are healthy, Kazakh/dark preferences and default viewport restored. Proof outside Git: home-order-ru-desktop.png (full page), home-order-ru-320.png and home-order-kk-360.png in the parent outputs/ui-review directory.

- Desktop: 1280px reading-site container, logo left, search/catalogue/translator destinations centrally, personal icons right. Mobile: compact search at the top and five inset bottom destinations.
- First content: heading, compact navigation chips, horizontal cover shelf. Approximately 146px covers on desktop and 116px on phones; full cards are native title links. Type/actual rating are placed over artwork; two-line names below. No slogans or metadata paragraphs.
- Catalogue navigation uses real updated/new/completed filters. No invented most-read ranking, time-period ranking, ratings, advertising blocks, messages or unrelated premium controls.
- Continuation retains real reading progress in linked illustrated rows, first for returning readers and after discovery for new visitors, using available catalogue covers without extra image-metadata requests. Unknown covers use a compact icon, not a fabricated image. Editorial shelf only exists when real editorial picks exist. Updates use clickable thumbnails and adjacent chapter buttons rather than wide text-only rows and detached controls.
- Existing weak-network image policy, Go media gateway, language preference and domain boundaries remain intact. Additional images are lazy except the first three covers. No new font, carousel library or remote image dependency.

## Scope

### Menu and choice controls follow-up

The owner's next correction explicitly includes the menu and selectors. Menu destinations now use icon-backed tiles/compact row buttons without explanatory paragraphs. Guest return paths and staff-only links remain intact. Language and theme use reusable native button groups with selected state in menu, settings and reader; other lists retain accessible native select behavior with consistent padding, themed arrow and44px targets. This deliberately avoids a JavaScript-heavy custom combobox for low-end Android.

The first phone review found an unwanted line break in Қазақша. Fixed the language group's width and chip label wrapping; Kazakh and Russian now fit at320px. Menu links/choices measure at least44px; there is no document-width overflow. Russian/light, Kazakh/dark and the reader's independent theme/select styling were checked. Language/theme switching persists after reload; catalogue type selection still submits its actual GET filter and returns the expected two manga titles.

Final frontend build passes all40 application/infrastructure tests,21 UI contracts, architecture/contrast gates, application and preview checks with zero diagnostics, and production compilation. Menu-specific tests preserve actual destinations, selected language, private-email omission, guest login return and reader/moderator/admin link visibility. Runtime backend contracts were not changed. Screenshots are saved outside Git as menu-controls-kk-360.png and menu-controls-desktop.png in the parent outputs/ui-review directory.

During this pass the owner reported WSL memory usage. Container measurements were about382MiB including the unused component-preview server; stopping only that tools-profile service reduced the running stack to about281MiB. The main site remains healthy. Windows still reported about3.4GiB for the WSL VM, while Linux reported about4.3GiB of reclaimable file cache after repeated builds; these different accounting figures must not be added or equated. No leak conclusion or load-test guarantee follows from an idle snapshot. No volumes, global WSL limits or other applications were modified. Avoid repeated full builds during small UI iterations, stop optional tools after use, and group final verification into one build.

This pass covers home, shared navigation, catalogue cover density and update presentation. It does not claim to reproduce every title/account/studio/reader screen from Senkuro. The small local catalogue remains real; artwork and additional titles are not copied or fabricated to fill eight desktop slots.

## Verification

UI contracts cover linked cover/release destinations, vote-gated scores, bilingual search submission, independently failed catalogue and usable updates, selected filter links and empty shelves. Infrastructure tests reject non-Go thumbnail URLs; the PostgreSQL test checks that update covers follow the title. Browser checks and final build evidence are appended after validation.

### Checked result

- Final Alpine web build passes: 40 application/infrastructure tests (193 assertions), 18 UI contracts (600 assertions), application/preview diagnostics with zero errors or warnings, import boundaries, both palette contrast checks and production bundle. Full Go suite with PostgreSQL and codec tests passes.
- Live desktop at1280px, Kazakh dark at360/320px and Russian light at320px inspected. At320px document width is305px with the browser scrollbar; filter chips fit without internal scrolling. Measured visible home actions are at least44px tall.
- All six current home images decode through Go. First covers are at197px from the top at360px; the chapter section starts at453px instead of preceding the covers.
- Carousel arrow changes scroll position and enabled state. Updated/new/completed navigation changes the actual route/data; completed currently shows an honest empty result. Cover opens its title; direct chapter button opens the real reader and the owned synthetic page decodes at800×1200.
- Search submits a real catalogue query, closes the native dialog on navigation, returns one expected local result, and Escape restores trigger focus. The initial manual test found the layout-persisted modal remained open after SvelteKit navigation; submit/catalogue-link handlers now close it and the flow was retested.
- Restored Kazakh/dark preferences and default viewport. Existing user home tab refreshed. Local containers healthy; production untouched.
- Screenshots outside Git: outputs/ui-review/senkuro-layout-kk-360.png, senkuro-layout-ru-light-320.png and senkuro-layout-desktop.png in the parent task directory.

The local catalogue has only three test titles, two using the owner's same test cover. This does not visually approximate a mature eight-title catalogue's variety. No new titles/artwork were fabricated. Physical weak Android performance and a populated continuation-card visual state were not established by these checks; continuation selection's existing unit contracts pass.

Local-only owner follow-up,23 October2026: latest updates initially show six compact rows, then reveal six more per explicit Load more action. A visible loaded-item count distinguishes the fetched page from the complete feed (a trailing + indicates a next cursor); All updates retains the full paginated route. After expansion keyboard focus moves to the first newly revealed chapter. Continuation remains first for returning readers and shows a labeled real chapter percentage beside its thin progress track; long team metadata stays on one line. No backend, data or Admin changes. Owner acceptance and push remain pending.

Cover badges share one nonwrapping top row: demo at the left, personal reading status beside it, optional age classification retained. Long status labels truncate visually with full accessible text; they do not force a second row. Local review only, no push.
