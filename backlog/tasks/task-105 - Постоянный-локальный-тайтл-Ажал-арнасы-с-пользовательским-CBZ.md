---
id: TASK-105
title: Постоянный локальный тайтл Ажал арнасы с пользовательским CBZ
status: Done
assignee:
  - '@codex'
created_date: '2026-10-20 20:04'
updated_date: '2026-10-20 20:09'
labels: []
dependencies: []
type: enhancement
ordinal: 47005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец просит постоянно загрузить локально Chapter-1-Kazakh-v2.cbz и данные указанного production-тайтла. Только localhost18080; публичные метаданные/обложка,54пользовательские страницы, обычная обработка и локальная публикация, сохранение существующих volumes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Локальный тайтл имеет метаданные и обложку указанной страницы; глава содержит54обработанные страницы из CBZ и открывается в читалке.
- [x] #2 Данные переживают перезапуск сервисов; production и существующие локальные тайтлы не изменены, повтор запуска не создаёт дубликат.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Проверить архив, создать отдельный локальный тайтл/проект, импортировать публичные metadata/artwork через существующие API, upload/process54pages, проверить reader и сохранение после restart; записать локальные URLs.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Completed local-only permanent import: /titles/ajal-arnasy, reader /read/4fa60dd5-ee9d-4315-9793-83c96319ee55 at localhost18080. Public metadata/cover from requestedproductiontitle copied through normaladmin/moderated-artwork API; localratings/comments remainlocal, no productionuser data copied. CBZ113MiB exceedsUI100MiB; boundedlocal extraction validated54PNG/112MiB,total250MiB/perimage25MiB cap, naturalorder andZIPCRC. All54uploaded through existingauth API/worker,ready,publishedlocally viaattribution/submit/moderation. No productlimits or productionmodified. Independentreview confirmed54orderedmanifestpages and realWebP1/27/54. Restarted PostgreSQL/MinIO and appservices with SAMEproject/credentials/volumes: title/cover/chapter/54orderedpages/sampleWebP verified again; all7serviceshealthy. Completed-import rerun observed existingpublishedtitle/no duplicate. One-time importer has partial-failure resume limits, not advertised asgeneralrestoretool. Permanentprivateconfig ~/.config/qmanga/local-review/stack.env0600 in0700directory, localaccounts/receipt retainedprivately; old tmp/task68/fixture.json moved out of automatedfixtures and persistentuserdata marker added. docs/LOCAL-DEVELOPMENT.md documents manualreviewenvironment and startup/no-volume-deletion. OriginalCBZunchanged; images/credentials notinGit. Actualbrowserreader and titleopened; localtitle screenshot ignoredtmp/local-library/local-title.jpg. Localdataonlytask fullyaccepted; no productionrelease needed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Постоянно импортированы локально Ажал арнасы, обложка и глава из54страниц. Читалка/WebP, повтор без дубликатов и сохранение после перезапуска PostgreSQL/MinIO проверены; независимый review пройден. Production не менялся; постоянный приватный конфиг и запуск документированы.
<!-- SECTION:FINAL_SUMMARY:END -->
