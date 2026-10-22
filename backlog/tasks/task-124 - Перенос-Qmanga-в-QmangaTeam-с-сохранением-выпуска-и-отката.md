---
id: TASK-124
title: Перенос Qmanga в QmangaTeam с сохранением выпуска и отката
status: Done
assignee:
  - '@codex'
created_date: '2026-10-22 19:46'
updated_date: '2026-10-22 20:28'
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
- [x] #2 CI публикует валидные lowercase GHCR references после смены владельца; проверки и независимый review пройдены
- [x] #3 Локальный и серверный origin обновлены; текущие и предыдущие immutable images доступны для выпуска и отката; production SHA и health проверены
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Завершить согласование базовых прав и перенос через Chrome. 2. Исправить GHCR references обоих workflows и подготовить миграцию namespace без потери active/previous tags. 3. Локальные проверки и независимый review, отдельный проверенный коммит; push только после разрешения владельца. 4. Наблюдать полный Verify и штатный выпуск, обновить origin и проверить production/rollback tags. Не менять SSH/Oracle права и не обходить deployment gates.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Владелец разрешил перенос, полный GitHub-доступ maqsatto/Bauka07, базовую Read, удаление старой приватной копии 0792a44 после отдельного подтверждения, push TASK-124 и public для трех migrated packages с последующим возвратом запрета. Bauka нужен только GitHub; SSH/Oracle-права не расширялись.

GitHub: Qmanga Team / QmangaTeam; рабочий репозиторий QmangaTeam/qmanga; maqsatto и Bauka07 Owner, Ygrik-CS Member, базовая Read. Три packages qmanga-api/web/gateway Public; org public package creation=false, private=true, internal=false, inherit_access=true после повторного открытия настройки.

Код: 7a09cf1 — lowercase GHCR references и regression в общем CI; 1722e83 — manual-only bootstrap/migration immutable images с уникальным bootstrap tag и сверкой raw manifest SHA256 без ослабления fail-closed guards; 90b6db03f59dfaf1a19daa40e1cb7ad0d4bf7c6b — новые default prefixes lib/Compose/env и актуальные адреса. Независимые review каждого существенного пакета и финальный review без блокеров.

Локальные проверки: red: 2 fail до исправления; green: 2 pass / 16 assertions после; bash -n, immutable aliases regression, internal staging route regression, OCI backup hook fixture, compose config и явный legacy override, backlog doctor, diff --check pass. Это fixture-проверка backup hook, не новый production restore drill.

Remote: Registry migration 39843716253 success; полный Verify 39844622844 и Fast Release 39844622864 success на 90b6db0, включая обе acceptance-группы и все 3 ARM64 publications. refs/deploy-ready=90b6db0. Реальные anonymous ARM64 pulls всех 5 current/previous refs на VM с пустым Docker config pass; все новые image IDs равны исходным. Старые packages сохранены.

Live acceptance 2026-10-23: staging и production active blue 90b6db03f59dfaf1a19daa40e1cb7ad0d4bf7c6b, previous green 1722e837c4d3ad9011cf9479e854c10cd45ba30c. Обе главные содержат exactSHA; health PostgreSQL/S3/status ok. Все4production Config.Image из нового namespace, actual image IDs совпадают с manifest-tag IDs. Staging 90 сохраняет старые ref names из уже загруженного lib первого тика; actual IDs всех 4 сервисов совпадают с новым namespace. Это проверенное переходное состояние до следующего обычного релиза, не необходимость ручного перезапуска. Previous manifest images доступны в обоих namespaces с одинаковыми IDs; реальный rollback не выполнялся. Host acceptance log: tmp/task-124/host-verification.log (ignored).

Host: config/staging.env и production.env атомарно изменены только по IMAGE_PREFIX; mode 0600 сохранен, секреты не выводились. Фактических systemd prefix overrides нет, timer active, hold текущего SHA нет. Local origin git@github.com:QmangaTeam/qmanga.git работает существующим SSH-ключом maqsatto; VM origin https://github.com/QmangaTeam/qmanga.git и fetch работают. Старый HTTPS OAuth push не имел workflow scope; решено существующим SSH, scopes не расширялись. Обновлены текущие ссылки и реальный checkout path в OPERATIONS/RUNBOOK.

Действий владельца для завершения TASK-124 не осталось. Сохранять старые packages и прежние immutable tags для переходного staging и отката.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Организация и рабочий репозиторий перенесены в QmangaTeam, maqsatto/Bauka07 имеют Owner, базовая Read. Мигрированы и проверены immutable ARM64 images и rollback refs; public-настройка трех packages завершена с возвратом запрета создания новых public packages. Новый CI и production используют новый registry; local/VM origins обновлены. Полный Verify и Fast Release 90b6db0 success, оба окружения автоматически выпущены на 90b6db03f59dfaf1a19daa40e1cb7ad0d4bf7c6b; внешние SHA/health и фактические image IDs проверены. Переходные staging ref names сохранены с доказанно идентичными IDs. Независимый review без блокеров; оставшихся действий пользователя нет.
<!-- SECTION:FINAL_SUMMARY:END -->
