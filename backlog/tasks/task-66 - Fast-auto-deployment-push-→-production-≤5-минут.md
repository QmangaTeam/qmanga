---
id: TASK-66
title: 'Fast auto deployment: push → production ≤5 минут'
status: In Progress
assignee: []
created_date: '2026-10-20 01:19'
updated_date: '2026-10-20 02:49'
labels:
  - CI
  - deployment
dependencies: []
priority: high
ordinal: 41026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Реализован быстрый автодеплой (коммиты 1f6c84a..5c98ee8): Fast Release (release.yml) — селективная сборка компонентов (GHA cache) относительно refs/deploy-ready, Full Verification отдельно и не блокирует fast path; slow path для migrations/deploy/infra/.github ждёт Verify. VM poller (timer 20s после обновления юнита владельцем) читает deploy-ready через git, staging smoke → production автоматически. release.sh: manifest-теги (неизменённые компоненты переиспользуются; production промоутит manifest staging; отсутствующий образ → предыдущий рабочий тег), backup skip только при свежей проверенной копии ≤30h и без миграций, auto-rollback + deployment-hold, проверка SHA в публичном футере. web.Dockerfile без тестов; фронтенд-проверки остались в Full Verification. Измерено (production, 5-мин таймер): CSS f7b0ffb 350s (Fast Release 59s), Svelte abd9e98 300s, Go 8d7d2dd 351s (Fast 102s); drill health-fail 504d3ab: production остался 8d7d2dd, плохой SHA ушёл в hold. Найден и исправлен баг: пустой net-diff реверта выбирал несуществующие теги api/gateway. Осталось: обновить systemd timer/service на VM (sudo install ... + daemon-reload + restart), двойной push, финальный отчёт/доки.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 push в main → production ≤300с без ручных действий
- [ ] #2 селективная сборка компонентов относительно последнего выпущенного SHA
- [ ] #3 backup при миграциях; иначе fresh-check ≤30h
- [ ] #4 авто-rollback при провале health; три контрольных релиза замерены
- [ ] #5 owners review: независимый субагент
<!-- AC:END -->
