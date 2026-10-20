---
id: TASK-66
title: 'Fast auto deployment: push → production ≤5 минут'
status: In Progress
assignee: []
created_date: '2026-10-20 01:19'
updated_date: '2026-10-20 01:24'
labels:
  - CI
  - deployment
dependencies: []
priority: high
ordinal: 41026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Переработка CI/CD: быстрый release path (селективная сборка изменённых компонентов, publish в GHCR, git-ref deploy-ready) + неблокирующая Full Verification (verify.yml). VM-poller каждые 20-30с по deploy-ready, staging→production автоматически. Backup из критического пути при отсутствии миграций заменён проверкой свежести (≤30h); при миграциях полный backup. web.Dockerfile без тестов (только bun run build). План: docs/superpowers/plans/2026-10-20-fast-auto-deployment.md
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 push в main → production ≤300с без ручных действий
- [ ] #2 селективная сборка компонентов относительно последнего выпущенного SHA
- [ ] #3 backup при миграциях; иначе fresh-check ≤30h
- [ ] #4 авто-rollback при провале health; три контрольных релиза замерены
- [ ] #5 owners review: независимый субагент
<!-- AC:END -->
