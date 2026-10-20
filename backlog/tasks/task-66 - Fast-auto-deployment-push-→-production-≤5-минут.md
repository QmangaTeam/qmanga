---
id: TASK-66
title: 'Fast auto deployment: push → production ≤5 минут'
status: Review
assignee: []
created_date: '2026-10-20 01:19'
updated_date: '2026-10-20 03:43'
labels:
  - CI
  - deployment
dependencies: []
priority: high
ordinal: 41026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Реализован и проверен автоматический путь push → staging → production (1f6c84a..36b3a5f). Fast Release (release.yml) классифицирует изменения относительно refs/deploy-ready, собирает только затронутые компоненты с GHA cache и двигает ref; Full Verification (verify.yml) идёт отдельно и не блокирует fast path, slow path (backend/migrations, deploy, infra, .github) ждёт Verify того же SHA. VM-poller читает deploy-ready через git, выпускает staging и на следующем тике production. release.sh: тег каждого компонента = последний коммит ≤ SHA, менявший его входы сборки; отсутствующий образ — fail-closed с hold (без молчаливой подстановки старого компонента); сверка запущенных контейнеров с выбранными image ID; production промоутит manifest staging; backup пропускается только при проверенной копии ≤30h и без миграций; auto-rollback + deployment-hold; проверка SHA в публичном футере. Poller без re-exec (нет busy-loop/обрыва по TimeoutStartSec). web.Dockerfile без фронтенд-проверок; они остались в Full Verification. Проверено на production: CSS 350s, Svelte 300s, Go 351s, web+api 332s, slow path 685s/1012s, double-push 315s (финальный второй SHA), drill 504d3ab (production остался на прежнем SHA, плохой SHA в hold), docs-only push 5dacbbd (0 workflow-запусков, ref не сдвинут), Verify 36b3a5f/a06546e зелёный. Два независимых review субагентом: CI-пайплайн (blocker по выбору тега исправлен в a06546e) и UI-фикс TASK-67. Осталось: владельцу установить обновлённый systemd service/timer (команда в OPERATIONS.md), затем подтвердить ≤300s новыми замерами; миграционный релиз этим путём ещё не прогонялся.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 push в main → production ≤300с без ручных действий
- [x] #2 селективная сборка компонентов относительно последнего выпущенного SHA
- [ ] #3 backup при миграциях; иначе fresh-check ≤30h
- [x] #4 авто-rollback при провале health; три контрольных релиза замерены
- [x] #5 owners review: независимый субагент
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Блокер: SSH к VM из окружения агента не проходит (Connection timed out), поэтому 20-секундный таймер не установлен — сейчас действует старый 5-минутный, из-за чего замеры 300–1012s. После установки юнита ожидается ~130–260s. Push выполняется через SSH-агент.
<!-- SECTION:NOTES:END -->
