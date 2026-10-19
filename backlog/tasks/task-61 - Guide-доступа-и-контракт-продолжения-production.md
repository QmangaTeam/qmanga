---
id: TASK-61
title: Guide доступа и контракт продолжения production
status: Done
assignee:
  - '@codex'
created_date: '2026-10-19 01:51'
updated_date: '2026-10-19 02:00'
labels: []
dependencies: []
ordinal: 36026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец просит после выпуска подготовить для следующего агента Guide: адреса, ссылки на ключи, подключение к серверу, добавление изменений и production контракт; закрепить в AGENTS.md. Расширить существующий docs/deployment/OPERATIONS.md вместо дублирующего отчёта. Только пути и команды, без значений секретов.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Guide содержит проверенные SSHпути, адреса, окружения, CI/staging/prod, backup/rollback и handoff
- [x] #2 AGENTS.md ссылается на Guide и закрепляет обязательный контракт; ссылки и команды сверены с кодом
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
После наблюдаемого prod18eb4fd подготовлен существующий OPERATIONS.md: проверенные SSHадрес/пути/права, серверные каталоги, CIimages/poller/sameSHA, backup/rollback, секреты только по путям, Googlesettingslink и handoff. AGENTS.md закрепляет контракт и ссылку; RUNBOOK исправлен под актуальный Googleworkflow и browsergroups. Команды сверены с deploy/lib.sh/release.sh/compose/verify.yml; git diff --check успешен.
<!-- SECTION:NOTES:END -->
