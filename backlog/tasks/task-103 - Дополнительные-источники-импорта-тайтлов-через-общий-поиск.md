---
id: TASK-103
title: Дополнительные источники импорта тайтлов через общий поиск
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 16:49'
updated_date: '2026-10-20 17:43'
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
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
AniList extends the same source picker, metadata form and tag mapping contract alongside verified Shikimori .io. Live search/details/covers and source-isolated fixtures passed; missing unsupported values stay unknown. Ready locally for release acceptance.
<!-- SECTION:FINAL_SUMMARY:END -->
