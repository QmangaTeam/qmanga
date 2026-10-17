# Senkuro layout adaptation

Owner direction, 18 October 2026: copy Senkuro's placement and interaction patterns, retain Qmanga's palette, use less text and more clickable artwork and recognizable action controls. This replaces the rejected text-first home composition, not the backend workflow contract.

## Visual contract

- Desktop: 1280px reading-site container, logo left, search/catalogue/translator destinations centrally, personal icons right. Mobile: compact search at the top and five inset bottom destinations.
- First content: heading, compact navigation chips, horizontal cover shelf. Approximately 146px covers on desktop and 116px on phones; full cards are native title links. Type/actual rating are placed over artwork; two-line names below. No slogans or metadata paragraphs.
- Catalogue navigation uses real updated/new/completed filters. No invented most-read ranking, time-period ranking, ratings, advertising blocks, messages or unrelated premium controls.
- Continuation retains real reading progress after the initial shelf in linked illustrated rows, using available catalogue covers without extra image-metadata requests. Unknown covers use a compact icon, not a fabricated image. Editorial shelf only exists when real editorial picks exist. Updates use clickable thumbnails and adjacent chapter buttons rather than wide text-only rows and detached controls.
- Existing weak-network image policy, Go media gateway, language preference and domain boundaries remain intact. Additional images are lazy except the first three covers. No new font, carousel library or remote image dependency.

## Scope

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
