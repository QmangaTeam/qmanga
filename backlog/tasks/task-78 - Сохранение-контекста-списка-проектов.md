---
id: TASK-78
title: Сохранение контекста списка проектов
status: In Progress
assignee:
  - '@studio-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 16:33'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/publishing/presentation/Projects.svelte
priority: high
type: enhancement
ordinal: 44010
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №10. Владелец: «добавь таски и начни исполнять».

После открытия проекта и возврата переводчик должен продолжать прежний поиск и выбор команды.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Применённые поиск и команда восстанавливаются через URL/возврат; редактируемый ввод не подменяет применённые критерии.
- [ ] #2 Позиция списка и дозагрузка восстановлены безопасно; удалённый доступ/смена аккаунта не восстанавливают чужие данные.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить контракт и добавить регрессии. 2. Реализовать согласованное изменение без новых правил и прав. 3. Выполнить проверки, записать SHA и ограничения приёмки.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
54b5b29: применённые q/team, число страниц и scroll в безопасной URL-ссылке возврата; повторная авторизованная загрузка до 20 страниц, generation/reset исключают поздние ответы старого аккаунта; черновой ввод не заменяет применённый q. Локально: svelte-check 0 ошибок/0 предупреждений; test:ui 46 pass; 9 domain/application тестов pass; frontend architecture pass; node --check studio-workflow.test.mjs pass. Browser studio-workflow.test.mjs/publishing-steps.test.mjs/project-cards.test.mjs и PostgreSQL integration на итоговом image ещё не выполнены; выпуск/независимая приёмка не подтверждены. Следующий шаг: общий rebuild и реальные browser/Go проверки координатором, затем Review; production SHA проверяет координатор.

Independent review P2 исправлен: 8414f46 сохраняет q/team/pages/scroll при возврате из preview; a34fe2b пишет ограниченный scroll в текущую history entry с throttle 150 ms и flush перед уходом/reload; guard owner/generation/path и запрет записи во время восстановления. Browser scroll→reload и полный возврат добавлены. Проверки после исправлений: svelte-check 0 errors/0 warnings; UI46 pass; 9 context/application тестов pass (54 assertions); architecture pass; browser syntax pass. Browser rerun на финальной сборке выполняет координатор перед новым independent review/push.
<!-- SECTION:NOTES:END -->
