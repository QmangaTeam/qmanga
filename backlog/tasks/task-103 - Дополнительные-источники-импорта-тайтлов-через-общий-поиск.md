---
id: TASK-103
title: Дополнительные источники импорта тайтлов через общий поиск
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 16:49'
updated_date: '2026-10-20 18:04'
labels:
  - web
  - publishing
  - imports
dependencies:
  - TASK-101
  - TASK-102
priority: high
type: feature
ordinal: 45005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Продолжение запроса «Shikimori и других сайтов». Расширить тот же поиск адаптерами других каталогов; первоначальные кандидаты MangaDex и AniList. Перед реализацией подтвердить их актуальные документированные способы доступа/доступные поля и зафиксировать поддерживаемые источники в этой карточке, без обхода ограничений сайтов.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 В том же поиске можно выбрать поддерживаемый внешний источник; результаты и внешний ID не смешиваются, название/тип/изображение помогают выбрать правильный тайтл.
- [x] #2 Как минимум один дополнительный источник помимо Shikimori работает через общий контракт заполнения метаданных и тегов, без отдельной несовместимой формы; недоступные поля остаются честно неизвестными.
- [x] #3 Кеш/таймаут/лимиты, provenance, обновление с сохранением ручных правок и ошибки реализованы отдельно для каждого адаптера; отказ одного источника не блокирует остальные, есть source fixtures и browser поиск→выбор→импорт.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Verify AniList official API contract and current rate limit. 2. Add AniList to shared search/selection UI and metadata mapping. 3. Verify isolated source errors, rate/cache constraints and fixtures.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Supported sources are Shikimori public REST API at https://shikimori.io and AniList public GraphQL API at https://graphql.anilist.co. Verified on 2026-10-20: Shikimori search, manga 13 details, author roles and actual cover all HTTP 200; .one timed out from this environment, so the adapter uses the reachable documented .io endpoint rather than shipping that known failure. Direct https://shikimori.io/api/doc/2.0 documentation confirms 5rps/90rpm and identifying User-Agent requirements; conservative per-source pacing is 1 request/sec. Official controller/profile serializer contracts also checked at https://github.com/shikimori/shikimori/blob/master/app/controllers/api/v1/mangas_controller.rb and app/serializers/manga_profile_serializer.rb. AniList official docs checked at https://github.com/AniList/ApiV2-GraphQL-Docs/blob/master/docs/guide/rate-limiting.md and docs/reference/object/media.md; current degraded limit is 30/min, adapter paces 2.5 seconds. Live full manga search and detail 30013 succeeded, including titles, synonyms, description, format/status, country, dates/year, cover, tags, genres and staff. Adapters share form fill and tag mapping contracts, have separate cache/budget/error states and honor upstream cooldowns; immediate selection waits ordinary local pacing without requiring user retry. Fixture regressions cover source errors, unsafe image hosts, partial data, cache reuse, actual search-thumbnail URL reuse and rate-limit propagation. Final coordinator rebuild/browser verification and release are still pending.

Final live common-port smoke against rebuilt localhost production stack: AniList search HTTP 200 with 10 results, explicit selection 30013 HTTP 200, 1 author, 62 canonical genre/tag mappings, 24 source candidates persisted, bounded/reprocessed WebP cover HTTP 200 (27896 bytes), anilist.co provenance. Shikimori independent live pipeline also passed in the same run. This uses actual upstream APIs plus local authorization, repository mapping/receipts and ImageMagick, not mocked source responses. Seven source fixture tests cover source-specific failure/cooldown/pacing/cache/partial data and thumbnail URL reuse. Final browser common-form scenarios pass on the rebuilt bundle. Latest code 6be46e4; integration, Verify migration gate and observed production release remain with coordinator.

Final integration 2026-10-20 at833f606: Docker API/web/gateway rebuilt and healthy. 158 frontend module +46 UI PASS; Svelte0errors/0warnings; architecture, both-theme contrast and preview checks PASS; Backlog doctor clean. Full Go/PostgreSQL -p1 ./... passed, latest catalog/source/HTTP/bootstrap real-DB checks repeated after endpoint/pacing changes PASS. Latest CBZ/PDF browser7/7 (including immediate cold decoder cancellation); metadata browser5/5. Root additionally verified12 responsive kk/ru x themes x320/360/1280 combinations for import forms and44px targets, and actual local API→live Shikimori.io/AniList search/detail/reprocessed WebPcover. Independent CBZ/PDF/metadata and root docs/test reviews have no remaining blocking findings. All12 new browser scenarios enter existing CI glob. Product tasks remain Review until full remote Verify and observed release of the same SHA. Production access blocker remains TCP22 timeout before authentication; no lock/hold/firewall/deployment bypass used.

Remote release readiness 2026-10-20: tested product SHA842a23f2c9bd55da1ad9447def70f428bda0b3b1. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39470308246 SUCCESS: both acceptance shards,88 browser scenarios (42+46),0 failures, full Go/PostgreSQL/codecs, frontend and all3 ARM64 images. Fast Release https://github.com/maqsatto/qmanga/actions/runs/39470308114 SUCCESS; refs/deploy-ready confirmed842a23f. Prior39468803806 failure was a real pre-hydration sidebar lost-Enter race, reproduced and fixed with native disabled-until-mounted control plus cold-JS regression; independent review passed, local workspace3/3/UI46/Svelte0/0 and repeated full CI green. Live Shikimori/AniList DTO→mapper→merge→metadataError/validPublication also validated (no titles created/changed). Public production still reports36b3a5f455eb37830408af44c2e2f00b48eabd51, health200; strict SSH with existing key and trusted known_hosts still times out before authentication. Rollout/production acceptance remains unconfirmed; keep Review. Next step: restore authorized SSH network access and inspect existing poller journal/state, preserve locks/backup/restore gates and release the ready immutable SHA through the single standard path.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
AniList extends the same source picker, metadata form and tag mapping contract alongside verified Shikimori .io. Live search/details/covers and source-isolated fixtures passed; missing unsupported values stay unknown. Ready locally for release acceptance.
<!-- SECTION:FINAL_SUMMARY:END -->
