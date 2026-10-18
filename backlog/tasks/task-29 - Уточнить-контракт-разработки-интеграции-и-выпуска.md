---
id: TASK-29
title: 'Уточнить контракт разработки, интеграции и выпуска'
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:26'
updated_date: '2026-10-18 21:35'
labels:
  - process
  - ci
dependencies: []
ordinal: 3026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец просит явный порядок разработки, интеграционных проверок и CI/CD. В проекте уже существуют Verify/Promote, host poller, release/rollback и runbook; нужно связать их с Backlog без второго процесса или дублирующих документов.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 В существующем runbook описаны границы изменений и проверяемые контракты UI/application/API/хранилища, локальные проверки, CI, staging, production и rollback.
- [x] #2 AGENTS.md ссылается на этот порядок; статусы Review и Done отражают фактическую приёмку и выпуск, результаты привязаны к SHA.
- [x] #3 CI проверяет целостность Backlog из зафиксированной версии инструмента, существующий полный набор интеграционных и браузерных проверок сохранён.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Сопоставить CI, Promote, host poller и runbook. 2. Дополнить существующий runbook контрактом выполнения задач и проверок на границах. 3. Добавить целостность Backlog в CI с pinned Bun/runtime и lockfile. 4. Проверить локально точную команду CI, ссылки и YAML; зафиксировать изменения.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Уточнён единый порядок в существующем RUNBOOK.md, AGENTS.md ссылается на него. Добавлена таблица контрактов UI/application/API/PostgreSQL/media/offline и матрица проверок, переходы Review/Done, привязка выпуска к SHA, staging и rollback. В Verify добавлены setup-bun 1.4.3 и locked install + backlog doctor; полный прежний CI сохранён, новые browser/*.test.mjs включаются автоматически. Локально выполнены точные shell-команды нового шага на Bun 1.4.3; doctor без ошибок. YAML и ссылки валидны; bash syntax, offsite backup checksum regression и shell setup recovery прошли. Начальный вариант запуска backlog в Alpine выявил glibc ELF несовместимость; итоговый workflow запускает CLI на Ubuntu runner через официальный setup-bun, не внутри Alpine. Remote Actions в этой сессии не запускались. Review до наблюдаемого CI для нового коммита.

Контракт и CI: cb0150c. Дополнительно реально выполнены полный Go/PostgreSQL/codec прогон и полный browser suite дважды; обнаруженная проблема тестовой изоляции оформлена TASK-30 и исправлена в 4b1b9af. Local evidence не заменяет ещё не наблюдавшийся Actions run для этих SHA.
<!-- SECTION:NOTES:END -->
