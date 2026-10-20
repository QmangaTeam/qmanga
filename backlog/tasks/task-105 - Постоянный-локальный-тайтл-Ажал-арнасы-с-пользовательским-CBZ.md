---
id: TASK-105
title: 'Постоянная локальная библиотека: отдельные тайтлы для CBZ и Ажал арнасы'
status: Done
assignee:
  - '@codex'
created_date: '2026-10-20 20:04'
updated_date: '2026-10-20 20:17'
labels: []
dependencies: []
type: enhancement
ordinal: 47005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Два разных локальных тайтла: Ажал арнасы с метаданными указанной production-страницы и Жұпар гүл асқақ жайқалады (Сака Миками) из CBZ. Имя и автор проверены на первой странице;54страницы относятся только ко второму тайтлу. Существующие volumes и ссылка читалки сохраняются.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Две отдельные публичные локальные карточки с собственными обложками; у Ажал арнасы нет чужой главы, у Жұпар гүл асқақ жайқалады одна глава из54страниц.
- [x] #2 Сохранены54обработанные страницы, прежний releaseID и постоянные volumes; production не изменён, инструкция запуска актуальна.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Проверить архив, создать отдельный локальный тайтл/проект, импортировать публичные metadata/artwork через существующие API, upload/process54pages, проверить reader и сохранение после restart; записать локальные URLs.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Completed local-only permanent import: /titles/ajal-arnasy, reader /read/4fa60dd5-ee9d-4315-9793-83c96319ee55 at localhost18080. Public metadata/cover from requestedproductiontitle copied through normaladmin/moderated-artwork API; localratings/comments remainlocal, no productionuser data copied. CBZ113MiB exceedsUI100MiB; boundedlocal extraction validated54PNG/112MiB,total250MiB/perimage25MiB cap, naturalorder andZIPCRC. All54uploaded through existingauth API/worker,ready,publishedlocally viaattribution/submit/moderation. No productlimits or productionmodified. Independentreview confirmed54orderedmanifestpages and realWebP1/27/54. Restarted PostgreSQL/MinIO and appservices with SAMEproject/credentials/volumes: title/cover/chapter/54orderedpages/sampleWebP verified again; all7serviceshealthy. Completed-import rerun observed existingpublishedtitle/no duplicate. One-time importer has partial-failure resume limits, not advertised asgeneralrestoretool. Permanentprivateconfig ~/.config/qmanga/local-review/stack.env0600 in0700directory, localaccounts/receipt retainedprivately; old tmp/task68/fixture.json moved out of automatedfixtures and persistentuserdata marker added. docs/LOCAL-DEVELOPMENT.md documents manualreviewenvironment and startup/no-volume-deletion. OriginalCBZunchanged; images/credentials notinGit. Actualbrowserreader and titleopened; localtitle screenshot ignoredtmp/local-library/local-title.jpg. Localdataonlytask fullyaccepted; no productionrelease needed.

Owner correction: CBZ belongs to a different title. First page visually inspected: Жұпар гүл асқақ жайқалады, Сака Миками. Correct association with separate local page and first-page cover; retain DeathTube metadata page. Preserve54uploadedpages and reader ID; verify both titles independently.

Correction supersedes original mistaken association. New title/project and first-page artwork created through local admin/moderated APIs. Guarded local PostgreSQL transaction moved chapter/release preserving revision/media. Source attribution corrected. Public APIs and independent review confirm old0/new1chapters,54orderedpages andnewtitleID. Protected receipt/docs updated; historical importer disabled. No production mutation.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Исправлена ошибочная привязка CBZ: Жұпар гүл асқақ жайқалады, Сака Миками имеет отдельную страницу, обложку с первого листа и54страницы. Прежняя ссылка читалки сохранена. Ажал арнасы оставлен отдельной карточкой без чужой главы. Независимый review пройден; только localhost.
<!-- SECTION:FINAL_SUMMARY:END -->
