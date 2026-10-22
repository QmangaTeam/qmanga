---
id: TASK-120
title: Перенос нового HTML-дизайна студии и согласование админ-панели
status: Review
assignee:
  - '@codex'
created_date: '2026-10-22 08:22'
updated_date: '2026-10-22 10:55'
labels: []
dependencies: []
priority: high
ordinal: 62005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенести qmanga-studio.html и экспортированный Figma-дизайн админ-панели на действующие страницы, сохранив реальные данные, права, загрузку/импорт, версии, модерацию и локальную проверку. Figma экспортирован через авторизованный браузер в .fig, 50-страничный PDF и SVG очереди; ограничение MCP не блокирует локальный образец.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Студия использует боковую навигацию, верхнюю строку и плотный список проектов из HTML, сохраняя реальные данные и ссылки.
- [x] #2 Редактор использует сетку страниц и компактные панели из HTML с рабочими загрузкой, порядком, сведениями, предпросмотром и отправкой.
- [x] #3 Админ-панель перенесена по доступному согласованному образцу без потери существующих функций.
- [x] #4 Проверены 320/360px, desktop, обе локали/темы, значимые сценарии и независимый review; результат доступен локально.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Локальная реализация студии и админки готова к дальнейшей приёмке владельцем, включая компактный экран команд и модальный просмотр админ-заявок. Исправлять конкретные замечания, сохранять права/точные ID/черновики. Push/deploy только по отдельному запросу.

Owner-approved project/editor mockup: 1. Project overview with internal sections and explicit chapter-version editor links; keep global Studio navigation. 2. Native new-chapter dialog and focused page/details/preflight editor using existing publishing operations. 3. Searchable taxonomy with limits and grouped complete title metadata form. 4. Mocked-write browser regressions, responsive/localized checks, independent review, small verified commits. Local only; no push/deploy.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-10-22: studio implementation committed a35ed6c (shell/directory), aa3cc39 (canvas/inspector). Fresh check 0 errors/0 warnings; 57 UI tests/824 assertions; 5 browser regressions (preflight, mocked reorder/failed attribution retry, mobile priority/empty project/shell) passed; separate 4-case sidebar matrix and admin all-nine-route 320/360/1280 both locale/theme matrix passed. Build, boundaries and palette contrast passed. Independent review found unnamed compact links; fixed and re-reviewed; final density CSS reviewed without findings. No push/deploy. Owner then enabled Figma account: exported full .fig (506994 bytes), 50-page PDF and queue SVG into Downloads through browser UI. Exact admin reference now available; proceed with admin implementation rather than await export. No credentials saved in repo.

2026-10-22: admin implementation committed 4f8805c (unified five-kind queue, exact revision canvas/decision, list-first catalogue, account/team/request/report inspectors, expandable audit, preserved nine routes) and dictionary follow-up (current preceding code commit). Final check 0 errors/0 warnings; 60 UI tests/843 assertions; fresh browser run 4/4 passed (nine-route responsive matrix kk/ru dark/light 320/360/1280; mocked page-two request deep link/reject retry; exact publication manifest/reject retry/publish target; dictionary dirty guard/overflow). Separate mocked metadata candidate regression 1/1 passed. Production build, architecture boundaries and theme contrast passed. Independent review found page-two queue-context bug; fixed with per-kind page depth and restored Teams paging, regression passed; subsequent admin and dictionary reviews have no actionable findings. Read loading no longer emits false Saved. Local admin browser opened with existing local admin account; Docker web runs Vite dev with live source/HMR. No real publish/approve/account mutations in browser tests. No push/deploy and no production SHA verification in this local-only package. Next: owner review of local admin/studio; adjust concrete feedback, then push/release only if requested.

Owner review follow-up: existing approved teams were not visible because original screen read only pending requests. Added approved-directory tab from existing admin/catalog-teams API, pending requests remain separate; real team avatars use existing public endpoint with fallback,56px list/88px inspector. Owner requested one local team application: created clearly labeled local UI demo, left pending, uploaded ownedQlogo for it. Public backend intentionally hides unapproved-team avatars; pending preview currently uses fallback (no access policy weakened). All admin API reads verified200:3approved/1pending teams,16titles,7accounts; audit50+; report/project/publication queues empty. Reviewer found approved-directory refresh and hash dirty-guard issues; both fixed and reviewed again without findings. New mocked browser regression passes directory/pending separation, avatar/fallback size, same-route deep selection, cancelled hash navigation retaining reason, and approve→directory refresh; no real approve/reject performed. Fresh check0/0 and UI60/843 pass; nine-route browser matrix and combined5browser scenarios passed before final two guard fixes; targeted team regression passed after fixes.

Code commits:4f8805c admin workspace,3b813e8 dictionaries; approved-team/avatar follow-up is the immediately preceding code commit. Final follow-up production build and architecture/contrast checks passed. Local review only; no production release or current production SHA verification.

Owner modal follow-up completed in7555cab: pending team Open request uses native showModal dialog, approved team details remain separate; all available request fields displayed (name/description/status/Telegram/review note/ID/slug). Native focus containment, close/Escape busy+dirty guards, return focus to trigger or pending tab, long mobile content scroll. Backend IDs/payloads unchanged. Regression first failed on missing dialog, then passed after implementation and focus-return fix. Fresh check0errors/0warnings,UI60/843,combined5browser scenarios,build,boundaries,contrast passed. Browser regression covers native :modal, fields,320/360/1280 dark/light bounds, long content, focus,close/Escape discard cancellation, busy rejection Escape,failed reject retaining reason/exact target,hash guard and approved directory refresh. Independent targeted review: no actionable findings. Test application remains pending; real admin decisions not performed. Opened separate local preview tab and captured screenshot under tmp/task120. No push/deploy.

studio/teams owner follow-up: removed duplicate intro card, single-team selector and irrelevant solo hint; one88px identity header with role/status, realTelegram and original project/request/profile/edit actions. Owner sees bounded member list; avatar settings, invitation history and invitation composer collapse independently. Counts remain unknown until successful load. Shared avatar picker now localized, keyboard-focusable and mobile-wrapping; same-actor successful save refreshes header image through versioned URL. Existing mutations, role constraints, stale team-response guards and paging retained. Fresh check0errors/0warnings,UI60tests/843assertions,7browser scenarios (studio3+admin4),production build,boundaries,contrast passed. Both locales/themes and320/360/1280 covered; mock invite failure preserves draft/exact team, team-switch confirmation, member-role restrictions and avatar-save refresh verified without real writes. Independent review and final invitation-wrapper review: no actionable findings. Updated legacy form/avatar-switch tests for intentional collapsed controls and missing single-team select; disposable legacy suites not run against owner data. Screenshot tmp/task120/studio-team-final.png. Local dev/HMR visible in owner tab; no push/deploy/current production SHA verification.

Studio-team compact UI code SHA:3429562. Owner review remains pending; statusReview.

Owner-approved project/editor mockup implemented locally in 611d8a2 (overview, internal project navigation, native create dialog, separate resume action), 1524151 (grouped complete metadata, searchable taxonomy with 12/24 limits preserving legacy values), and preceding editor commit (page-focus canvas/filmstrip, native version switch, exact revision failures, preserved list return and stale account guards). Fresh svelte-check: 0 errors/0 warnings; UI:61 tests/846 assertions; production build, dependency boundaries, palette contrast and git diff check pass. Combined browser run:13/13 pass across project navigation/native creation retry and success, account-change during delayed creation, editor preflight/reorder/failed save, exact reader-preview return, 720-option taxonomy and failed-submit draft retention; kk/ru, dark/light and 320/360/1280 plus focus 1440 tested. Separate non-destructive publishing-steps subset:2/2 passed. Screenshots inspected under ignored tmp/task120. All publishing mutations in the new and executed regressions were mocked; destructive legacy create/publish/cover suites were adapted for navigation but not run on owner data. Independent review completed, stale reads and dependent writes guarded, no open findings. Local Vite/HMR remains available; no push/deploy/current production SHA verification. Next: owner checks locally; release only on separate request.

Final editor/guard code SHA:5451a3a. Validation above ran against the same saved product sources before this commit; only acceptance notes changed afterwards.

2026-10-22 deployment authorized by owner. Pushed accumulated frontend package. Fast Release39765129545 succeeded for4e4b8d87ac01005883f5d4656992a4fe177aeaa6; VM poller automatically released staging and production to green same SHA, observed2026-10-22T10:54:32Z. External HTTPS health PostgreSQL/S3/status ok and homepage exact SHA confirmed. Staging browser smoke36 responsive views kk/ru dark/light320/360/1280,0runtime errors,guest Studio sign-in passed. Production browser smoke pending final result. Full Verify is not green yet: earlier runs revealed stale UI test assumptions (import disclosure, removed team overview, native admin dialog/mobile navigation, collapsed admin metrics). Test-only fixesca5b27d,4e4b8d8,99fd185 retain real assertions; local project-cards1/1,form-workflow3/3,metrics2/2 pass, independent review no open findings. Superseded verification runs were cancelled only after fixes were queued; latest full Verify39766638560 on99fd185 is queued/running and must be observed before claiming full CI success. Runtime UI code is unchanged by these test-only commits. Keep Review until remaining CI/owner private-workspace acceptance; no production account/data mutations performed.

Final external production browser smoke passed48/48 responsive views,kk/ru,dark/light,320/360/1280,public home/catalog/login/title and guest Studio sign-in;0runtime exceptions. Observed production SHA4e4b8d87ac01005883f5d4656992a4fe177aeaa6. Latest test-only99fd185 remains on automatic release/Verify path; no full-CI success claimed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Локально готовы студия и админка по HTML/Figma. Код:a35ed6c,aa3cc39,4f8805c,3b813e8,8b5b1a9. Проверки:check0/0,UI60/843,build,architecture,contrast;responsive/browser5 сценариев и отдельная регрессия последующих team-guard исправлений прошли. Независимый review завершён без оставшихся actionable замечаний. В браузере открыта локальная тестовая заявка, Vite dev/HMR работает. Ожидается приёмка владельцем. Публичный avatar endpoint скрывает изображения неодобренных команд; для заявки отображается fallback. Push/deploy не выполнялись.

7555cab: заявки команд открываются модальным окном с дополнительными доступными полями; проверено локально и независимо reviewed, ожидается приёмка владельцем.

Страница команд студии стала компактнее: единая шапка, видимые участники, раскрываемые настройки/приглашения и локализованный выбор аватара; проверки и независимый review пройдены, ожидается локальная приёмка.

Owner-approved project workspace, focused chapter editor and large searchable genre/tag chooser are implemented and checked locally. Code commits611d8a2,1524151 and latest editor commit; check0/0,UI61/846,browser13/13,build,boundaries,contrast passed. Independent review has no open findings. Review status retains owner local acceptance; no push or release.

Интерфейс выпущен: staging и production наблюдались на4e4b8d87ac01005883f5d4656992a4fe177aeaa6,Fast Release39765129545 success,HTTPS health/SHA и48 внешних responsive views прошли. Test-only99fd185 опубликован после исправления устаревших проверок; полный Verify39766638560 ещё нужно дождаться. Review сохранён из-за оставшегося CI и приёмки приватных рабочих областей владельцем.
<!-- SECTION:FINAL_SUMMARY:END -->
