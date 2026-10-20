---
id: TASK-66
title: 'Fast auto deployment: push → production ≤5 минут'
status: Review
assignee: []
created_date: '2026-10-20 01:19'
updated_date: '2026-10-20 18:19'
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

2026-10-20: owner explicitly requested immediate production rollout. Rechecked refs/deploy-ready = 4145910eb7bc649aaca2a2e890a44fbfa5f75575 with successful Fast Release and full Verify. Public DNS resolves VM 138.3.254.185; TCP443 connects, TCP22 times out before authentication. qmanga_deploy exists mode0600; trusted VM host entries exist in ~/.ssh/known_hosts (old documented qmanga_known_hosts absent), OPERATIONS corrected. Repeated strict SSH with the actual known_hosts also times out. Production public SHA remains 36b3a5f455eb37830408af44c2e2f00b48eabd51. No firewall/permissions/lock/hold changes or alternate deployment path used. Next: restore existing authorized SSH network access, inspect poller journal/state and allow the tested SHA through standard staging/production gates.

Remote release readiness 2026-10-20: tested product SHA842a23f2c9bd55da1ad9447def70f428bda0b3b1. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39470308246 SUCCESS: both acceptance shards,88 browser scenarios (42+46),0 failures, full Go/PostgreSQL/codecs, frontend and all3 ARM64 images. Fast Release https://github.com/maqsatto/qmanga/actions/runs/39470308114 SUCCESS; refs/deploy-ready confirmed842a23f. Prior39468803806 failure was a real pre-hydration sidebar lost-Enter race, reproduced and fixed with native disabled-until-mounted control plus cold-JS regression; independent review passed, local workspace3/3/UI46/Svelte0/0 and repeated full CI green. Live Shikimori/AniList DTO→mapper→merge→metadataError/validPublication also validated (no titles created/changed). Public production still reports36b3a5f455eb37830408af44c2e2f00b48eabd51, health200; strict SSH with existing key and trusted known_hosts still times out before authentication. Rollout/production acceptance remains unconfirmed; keep Review. Next step: restore authorized SSH network access and inspect existing poller journal/state, preserve locks/backup/restore gates and release the ready immutable SHA through the single standard path.

Oracle Console read-only diagnosis: owner signed in through the visible browser. VM qmanga is Running, expected public address and subnet, no VNIC NSGs. Associated security list permits TCP22 from exactly one previous operator /32; verified current external operator IPv4 does not match that allowlist. HTTP80/HTTPS443 are public as expected; no OCI rules have been saved or changed. Prepared a single-/32 replacement in the existing SSH ingress form and requested explicit owner approval before Save; host firewall and poller state are still uninspected. No personal source IPs or account credentials recorded here. Bounded public observation after ready842a23f did not observe rollout; public API import route remains404 and productionSHA36b3a5f. Next: owner decision on the proposed restricted SSH source replacement, then SSH read-only poller/backup/hold diagnosis before normal gated promotion.
<!-- SECTION:NOTES:END -->
