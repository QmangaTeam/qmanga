---
id: TASK-124
title: Перенос Qmanga в QmangaTeam с сохранением выпуска и отката
status: Review
assignee:
  - '@codex'
created_date: '2026-10-22 19:46'
updated_date: '2026-10-22 19:57'
labels: []
dependencies: []
ordinal: 66005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец поручил перенести рабочий maqsatto/qmanga в организацию Qmanga Team, сменить slug MaqsattoTeam на QmangaTeam, дать maqsatto и Bauka07 полный GitHub-доступ и сохранить работу Oracle VM. Уточнение владельца: Bauka нужен только GitHub; SSH/Oracle доступ ему не добавлять. Старый приватный QmangaTeam/qmanga с SHA 0792a44 удален после отдельного подтверждения владельца. Рабочий репозиторий уже перенесен; оба аккаунта Owner; базовая Read согласована и подтверждена. Пользователь разрешил push и завершение настройки выпуска.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Рабочий репозиторий перенесен в QmangaTeam/qmanga; maqsatto и Bauka07 имеют полный доступ; базовые права организации согласованы
- [ ] #2 CI публикует валидные lowercase GHCR references после смены владельца; проверки и независимый review пройдены
- [ ] #3 Локальный и серверный origin обновлены; текущие и предыдущие immutable images доступны для выпуска и отката; production SHA и health проверены
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Завершить согласование базовых прав и перенос через Chrome. 2. Исправить GHCR references обоих workflows и подготовить миграцию namespace без потери active/previous tags. 3. Локальные проверки и независимый review, отдельный проверенный коммит; push только после разрешения владельца. 4. Наблюдать полный Verify и штатный выпуск, обновить origin и проверить production/rollback tags. Не менять SSH/Oracle права и не обходить deployment gates.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-10-23: slug QmangaTeam и display name Qmanga Team подтверждены через Chrome; старая приватная копия 0792a44 удалена после отдельного подтверждения. Рабочий репозиторий перенесен, maqsatto/Bauka07 Owner, базовая Read подтверждена после согласования. Local и VM origin обновлены на https://github.com/QmangaTeam/qmanga.git; VM ls-remote HEAD/main и refs/deploy-ready успешен. Staging/production active b7c5578403a7bf0bce27a4860bbc22accb68c773, предыдущий dfa6096f9a1cbd196885b9ac024ae3277365d327; внешний health PostgreSQL/S3 ok. Подготовлена lowercase-нормализация GHCR в обоих workflows и regression в общем CI. Red: 2 fail до правки, green: 2 pass/16 assertions после; component-alias regression pass; backlog doctor, bash -n deploy/*.sh и git diff --check pass. Независимый review не нашел ошибок diff, выявил bootstrap-риск нового package: live docker manifest inspect нового namespace вернул denied. Не ослаблять immutable-tag guard; требуется явная bootstrap-публикация новых packages, сохранение active/previous tags и проверка публичного pull. Deploy prefix пока старый намеренно. Push и переключение namespace еще не выполнены.

Коммиты 7a09cf1 (lowercase GHCR + regression) и 1722e83 (manual registry migration) опубликованы существующим SSH-доступом maqsatto. HTTPS OAuth push отклонен из-за отсутствия workflow scope; scopes/credentials не расширяли. Registry migration Actions 39843716253 success: active и previous image tags скопированы с проверкой raw manifest SHA256; старые packages сохранены. Новые packages private; изменение public заблокировано org package-creation policy. Подготовлено временное разрешение public с возвратом запрета после трех packages; ожидается отдельное browser confirmation владельца. Локальный второй пакет переключает default prefix lib/Compose/env на ghcr.io/qmangateam/qmanga и актуальные ссылки документов. bash -n, compose --images (новый default и явный старый override), diff --check, backlog doctor pass; независимый review без code blockers. Перед push обязательны anonymous pulls всех active/previous tags. Фактический systemd IMAGE_PREFIX override и DropInPaths отсутствуют; production healthy b7c5578. Первый staging tick может наследовать старый prefix из уже загруженного lib, поэтому final проверка включает фактические Config.Image.

Владелец разрешил public для трех migrated packages с возвратом запрета. Через Chrome временно разрешено public package creation, qmanga-api/web/gateway переведены Public; запрет public creation сразу восстановлен и проверен повторным открытием Packages settings (public=false, private=true, internal=false, inherit_access=true). Anonymous manifest checks всех 5 active/previous tags успешны с пустым Docker config. На Oracle ARM64 VM реальные anonymous pulls всех 5 tags успешны; image ID каждого нового ref точно совпал со старым namespace. Второй пакет готов: default prefix lib/Compose/env новый, актуальные ссылки документов обновлены, legacy override работает. Повторно pass: 2 regression tests/16 assertions, bash -n, internal-staging-route, OCI backup hook fixture, immutable aliases fixture, compose config --quiet, diff --check. Независимый review второго пакета без блокеров при сохранении старых packages и проверке Config.Image после выпуска. Следующий шаг: опубликовать проверенный default-prefix commit, дождаться полного Verify и штатного poller release того же SHA; проверить actual SHA/health/Config.Image/rollback tag availability.
<!-- SECTION:NOTES:END -->
