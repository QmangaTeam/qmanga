# Component-first redesign: implementation and verification

Date: 17 October 2026. Implemented directly in `outputs/qmanga`, on `main`.
Code checkpoints: `a526377` (components), `013f76c` (shell and public/account screens), `46eb105` (remaining bilingual workflows and connected reading).

## Delivered

- Pure reusable controls, accessible fields, state messages, navigation, cover cards, chapter and library rows; separate development-only component preview.
- Existing green/dark/light palette retained. Compact catalogue-first home, two-column phone catalogue, five-destination mobile navigation and task-oriented account forms.
- Explicit Kazakh-first/Russian UI across public screens, identity, reading, saved chapters, library, community, notifications, teams, publishing and moderation. Content itself is not machine-translated.
- Per-layout locale context, cookie-backed SSR language and browser preference adapter. Language changes do not remount the reader. Application failures and reader/download notices use stable codes rather than translated exceptions.
- Existing saved-download records remain compatible without changing IndexedDB keys, stored files or checksums. Generic offline HTML stays Kazakh and contains no account data.
- Real continuation on home, including pagination past non-reading library entries. First-chapter action on title; chapter number/name search and chronological ordering, with stable server pagination.
- Bilingual help on reading, offline limits, permitted publishing and the existing moderation/report flow.
- Existing Go/Gin application boundaries, private MinIO, backend-only image delivery, WebP profiles and Alpine runtime retained.

## Fresh automated gates

`docker compose build web api` and the final `docker compose build web` exited 0. The final clean web build ran:

| Gate | Result |
| --- | --- |
| Application/domain/adapter Bun tests | 28 passed, 0 failed; 141 assertions |
| Actual Svelte SSR component/UI tests | 13 passed, 0 failed; 578 assertions |
| Application Svelte check | 0 errors, 0 warnings |
| Standalone preview Svelte check | 0 errors, 0 warnings |
| Frontend dependency boundary check | Passed |
| Token text/action/control contrast, both themes | Passed |
| Production SSR/client/prerender build | Passed |

`docker compose --profile tools run --rm go-tests` exited 0 with actual PostgreSQL integration and ImageMagick codec tests. Chapter search tests cover normalized Kazakh queries, invalid sort/length, literal wildcard escaping, published-only visibility, descending cursors and distinct fractional chapter ordering.

New continuation tests were observed failing before implementation and passing afterwards. New help SSR test likewise failed before its component existed. Clean-build testing caught and corrected Kit alias initialization order and offline prerender access to query parameters; the final Docker gate includes these paths.

The default API, web, worker, gateway, PostgreSQL, MinIO and local Mailpit services were healthy after deployment. No storage volumes were removed.

## Browser checks actually performed

Browser emulation, not a physical Android:

- Component preview at 360px: Russian/light controls, long labels, minimum 44px buttons, no horizontal overflow; counter and busy interaction observed without duplicate callback.
- Settings at 320px: Russian labels, no horizontal overflow; measured selects/actions at least 44px high. At 360px: Kazakh/dark and Russian/light screenshots, persistent explicit language after reload, matching HTML language.
- Desktop help at 1280px: actual Russian sections present, no horizontal overflow.
- Title: first-chapter action resolves to a real release; chapter query and descending sort persist in the URL; unmatched query shows a localized empty state, matching number returns the chapter.
- Test reader account: login, save library entry, load discussion/rating controls, read a page, observe cloud-save acknowledgement, return home and see actual continuation.
- Reader: right edge advances; ArrowRight works after button focus; 1/5 → 2/5 → 3/5 on desktop. Enabled edge has pointer cursor; no browser errors were reported in that reader check.
- Reader language switch at 3/5 preserves release and page. Localized settings and progress notices update in place.
- Saved chapter: 0.7 MB test release reaches ready; open explicit offline reader, disable network, reload, restore page 3/5 and advance to 4/5. The displayed image is decoded from a local blob. Network emulation was restored afterwards.
- SSR HTTP requests with Russian cookie: home/catalogue return `lang=ru` and `private, no-store`; offline shell remains `lang=kk`. Production `/ui-preview` returns 404.

Screenshots are outside Git in the adjacent `outputs/ui-review/`: `settings-kk-mobile.jpg` and `settings-ru-light-mobile.jpg`. They contain neither credentials nor source manga artwork.

## Independent review

Read-only reviewer found two issues: continuation only inspected the first library page; home omitted the demonstration marker. Both were fixed and the reviewer rechecked the additions without substantive remaining findings. The multi-team first-read action also clears stale chapter filters. Reviewer did not independently rerun browser tests; the results above belong to the implementing agent.

## Explicit limits

- The exhaustive original browser matrix is not fully executed: no physical weak Android, true rural connection, 200% desktop browser zoom, on-screen keyboard, OS font scaling or long-session/backgrounding acceptance.
- No claim that every role mutation was repeated in the browser after redesign. Existing full Go authorization/workflow tests pass; fresh browser checks above identify the exact exercised paths.
- No production deployment or external mail delivery. Local Mailpit only; public TLS/domain, trusted-proxy per-client rate limiting, backups/restore, monitoring and CI remain launch work.
- No public rights-contact address was invented. Help points to the existing verified-account report flow; an external contact and content/age policy need to be established before public launch.
- Existing bounded studio/admin lists, profiles, advanced library filters and an actual chronological release feed remain backlog, not silently claimed delivered. Home shows recently updated titles, not a fabricated release feed.
- Passing text-token contrast checks is not a complete accessibility audit or proof of low-end-device performance.
