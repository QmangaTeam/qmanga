---
id: TASK-66
title: 'Fast auto deployment: push → production ≤5 минут'
status: In Progress
assignee: []
created_date: '2026-10-20 01:19'
updated_date: '2026-10-20 01:32'
labels:
  - CI
  - deployment
dependencies: []
priority: high
ordinal: 41026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Переработка CI/CD (реализовано, ждёт live-приёмки): Fast Release (release.yml) — селективная сборка компонентов с GHA cache и refs/deploy-ready; Full Verification (verify.yml) не блокирует обычный релиз; медленный путь для migrations/deploy/infra/.github. poll-staging.sh читает deploy-ready через git каждые 20с (таймер VM обновить владельцу: sudo cp deploy/qmanga-staging-poll.timer /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl restart qmanga-staging-poll.timer), затем staging smoke → auto production. release.sh: per-component immutable tags (manifest), backup skip при свежей копии ≤30h и без миграций, проверка SHA в публичном футере, auto-rollback сохранён. web.Dockerfile: только bun run build (решение владельца). Удалены SSH deploy job и promote.yml. Коммиты: 1f6c84a, 43c1a5d. План: docs/superpowers/plans/2026-10-20-fast-auto-deployment.md. Осталось: 3 контрольных релиза с замерами, health-fail rollback, double-push, обновление таймера на VM.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 push в main → production ≤300с без ручных действий
- [ ] #2 селективная сборка компонентов относительно последнего выпущенного SHA
- [ ] #3 backup при миграциях; иначе fresh-check ≤30h
- [ ] #4 авто-rollback при провале health; три контрольных релиза замерены
- [ ] #5 owners review: независимый субагент
<!-- AC:END -->
