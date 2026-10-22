---
id: TASK-124
title: Перенос Qmanga в QmangaTeam с сохранением выпуска и отката
status: In Progress
assignee:
  - '@codex'
created_date: '2026-10-22 19:46'
updated_date: '2026-10-22 19:49'
labels: []
dependencies: []
ordinal: 66005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец поручил перенести maqsatto/qmanga в организацию Qmanga Team, сменить slug MaqsattoTeam на QmangaTeam, дать maqsatto и Bauka07 полный GitHub-доступ и сохранить работу Oracle VM. Bauka нужен только GitHub, SSH/Oracle доступ ему не добавлять. Старый приватный QmangaTeam/qmanga с SHA 0792a44 удален после отдельного подтверждения владельца. Перенос рабочего репозитория и политика базовых прав пока ожидают подтверждения.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Рабочий репозиторий перенесен в QmangaTeam/qmanga; maqsatto и Bauka07 имеют полный доступ; базовые права организации согласованы
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
<!-- SECTION:NOTES:END -->
