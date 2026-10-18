---
id: TASK-16
title: Компактный футер и полезные ссылки
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
labels:
  - stage-4
dependencies: []
references:
  - UPDATE.md
ordinal: 16
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-4. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Футер: компактные колонки «Qmanga», «Читателям», «Переводчикам», «Қосымша / Дополнительно». Перенести туда «Настройки», «Мои обращения», «Помощь»; оставить работающие маршруты. Для переводчиков дать короткий маршрут «команда → заявка на проект → главы → модерация» со ссылками на настоящую справку, а не длинный текст в навигации.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Implement approved task criteria in existing architecture; add behavioral regressions, verify Docker integration and keep Review until CI/release. Legal/contact approval remains explicit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Compact bilingual four-column footer (two columns on phone) with working settings/reports/help routes, translator team/request/chapters/help#publishing chain, original legal draft routes and2026copyright. Version defaults honest local; deployment exposes exact RELEASE_SHA viaBUILD_VERSION. Native44px links and locale SSR test; actual browser locales/themes320/360/desktop passes. Remote acceptance/release pending.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
