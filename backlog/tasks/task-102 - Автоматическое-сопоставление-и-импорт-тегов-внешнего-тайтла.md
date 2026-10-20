---
id: TASK-102
title: Автоматическое сопоставление и импорт тегов внешнего тайтла
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
priority: high
type: feature
ordinal: 45004
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Часть поручения владельца: «и теги автоматом тоже». Единая политика для Shikimori и следующих источников; ID/slug существующей таксономии не подменяются внешними строками.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Теги/жанры выбранного тайтла автоматически сопоставляются с существующими taxonomy ID по внешнему стабильному ID и поддерживаемым алиасам; повторный импорт не создаёт дубликаты и не меняет смысл фильтров.
- [x] #2 Однозначные соответствия сразу заполняют выбор тегов в форме. Ненайденные теги автоматически сохраняются как кандидаты с источником для предусмотренного согласования/создания, не теряются молча; неоднозначные совпадения явно различимы и не выбираются наугад.
- [x] #3 Жанры, категории и группы сохраняют принятую модель; импорт не переводит автоматически собственные названия манги, не перезаписывает ручные правки без согласия и не обходит права/модерацию. Проверены разные языки, повтор, частичные данные и отказ источника.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Map stable external IDs and normalized taxonomy aliases without arbitrary ambiguous matches. 2. Persist unmatched/ambiguous source candidates for administrative review; retain canonical taxonomy IDs. 3. Verify repeats, language aliases and partial imports.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Canonical matching uses stable source IDs, normalized exact taxonomy Kazakh/Russian names and slugs (English slug aliases), explicit admin-reviewed mappings and deduplication. Unknown and ambiguous tags are persisted in metadata_tag_mappings; ambiguity never picks an arbitrary canonical ID. Admin taxonomy page offers one active candidate editor with explicit canonical genre/category selection, preserving taxonomy groups and existing IDs. Browser account/context changes invalidate pending results. Genre/category caps preserve existing selected tags and display omitted imported-tag counts instead of claiming a complete import. Added real PostgreSQL regression for repeated imports, changed source labels, unmapped/ambiguous candidates, permissions, explicit review and selection provenance; coordinator reports full PostgreSQL suite passed. Merge unit cases cover missing fields, repeats, caps and retained manual source links. Final browser candidate keyboard/save test awaits rebuilt bundle; independent review and production release remain with coordinator.

Final local acceptance: coordinator real-PostgreSQL suite passed the canonical mapping/candidate/review/provenance regression; source fixtures and 3 merge-policy tests pass. Rebuilt browser verifies canonical tags persisted through the actual admin save API and 100 candidates render no eager taxonomy selects, keyboard opens exactly one editor (<800 options), explicit canonical mapping is sent and the candidate leaves the queue. Actual Shikimori selection matched 4 canonical genres; actual AniList selection matched 62 and persisted 24 unresolved candidates. The existing 12-genre/24-category form limits remain authoritative: existing choices are retained and omitted imported-tag counts are visible. No duplicate taxonomy rows are auto-created. Independent review has no blocking findings. Code SHAs through 6be46e4; production acceptance remains pending coordinator release.

Final integration 2026-10-20 at833f606: Docker API/web/gateway rebuilt and healthy. 158 frontend module +46 UI PASS; Svelte0errors/0warnings; architecture, both-theme contrast and preview checks PASS; Backlog doctor clean. Full Go/PostgreSQL -p1 ./... passed, latest catalog/source/HTTP/bootstrap real-DB checks repeated after endpoint/pacing changes PASS. Latest CBZ/PDF browser7/7 (including immediate cold decoder cancellation); metadata browser5/5. Root additionally verified12 responsive kk/ru x themes x320/360/1280 combinations for import forms and44px targets, and actual local API→live Shikimori.io/AniList search/detail/reprocessed WebPcover. Independent CBZ/PDF/metadata and root docs/test reviews have no remaining blocking findings. All12 new browser scenarios enter existing CI glob. Product tasks remain Review until full remote Verify and observed release of the same SHA. Production access blocker remains TCP22 timeout before authentication; no lock/hold/firewall/deployment bypass used.

Remote release readiness 2026-10-20: tested product SHA842a23f2c9bd55da1ad9447def70f428bda0b3b1. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39470308246 SUCCESS: both acceptance shards,88 browser scenarios (42+46),0 failures, full Go/PostgreSQL/codecs, frontend and all3 ARM64 images. Fast Release https://github.com/maqsatto/qmanga/actions/runs/39470308114 SUCCESS; refs/deploy-ready confirmed842a23f. Prior39468803806 failure was a real pre-hydration sidebar lost-Enter race, reproduced and fixed with native disabled-until-mounted control plus cold-JS regression; independent review passed, local workspace3/3/UI46/Svelte0/0 and repeated full CI green. Live Shikimori/AniList DTO→mapper→merge→metadataError/validPublication also validated (no titles created/changed). Public production still reports36b3a5f455eb37830408af44c2e2f00b48eabd51, health200; strict SSH with existing key and trusted known_hosts still times out before authentication. Rollout/production acceptance remains unconfirmed; keep Review. Next step: restore authorized SSH network access and inspect existing poller journal/state, preserve locks/backup/restore gates and release the ready immutable SHA through the single standard path.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
External tags map by stable source IDs and exact normalized aliases to existing canonical IDs. Ambiguous/unmatched candidates persist for explicit admin agreement; one candidate editor and account/context guards keep the workflow bounded and safe. Local unit, PostgreSQL and browser checks pass; awaiting release acceptance.
<!-- SECTION:FINAL_SUMMARY:END -->
