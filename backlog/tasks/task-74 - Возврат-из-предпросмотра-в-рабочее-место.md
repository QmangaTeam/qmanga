---
id: TASK-74
title: Возврат из предпросмотра в рабочее место
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
  - apps/web/src/lib/features/publishing/presentation/PublishingStudio.svelte
priority: high
type: enhancement
ordinal: 44006
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №6. Владелец: «добавь таски и начни исполнять».

После проверки главы переводчику нужно продолжить ту же версию и шаг без поиска позиции заново.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Возврат из предпросмотра восстанавливает проект, версию, рабочую секцию и позицию.
- [ ] #2 Прямые ссылки, Back/reload и недоступная версия имеют честный fallback; приватность preview и account isolation сохраняются.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить контракт и добавить регрессии. 2. Реализовать согласованное изменение без новых правил и прав. 3. Выполнить проверки, записать SHA и ограничения приёмки.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
9c3bac5: preview URL несёт проект, ревизию, шаг и scroll; workspaceReturn принимает только точную локальную ссылку той же версии. Reader banner интегрируется владельцем Reader. Недоступная версия показывает явный fallback. Локально: svelte-check 0 ошибок/0 предупреждений; test:ui 46 pass; 9 domain/application тестов pass; frontend architecture pass; node --check studio-workflow.test.mjs pass. Browser studio-workflow.test.mjs/publishing-steps.test.mjs/project-cards.test.mjs и PostgreSQL integration на итоговом image ещё не выполнены; выпуск/независимая приёмка не подтверждены. Следующий шаг: общий rebuild и реальные browser/Go проверки координатором, затем Review; production SHA проверяет координатор.

Independent review P2 исправлен: 8414f46 сохраняет валидированный projectsReturn внутри workspace returnTo с запретом постороннего пути/схемы/account scope; df704ba синхронизирует выбранные revision/step через SvelteKit replaceState после завершения начального восстановления, не вызывая navigation/dirty confirm. Unit nested return RED→GREEN; browser full list→workspace→preview→return→all projects и выбор B/review→reload добавлены. Проверки после исправлений: svelte-check 0 errors/0 warnings; UI46 pass; 9 context/application тестов pass (54 assertions); architecture pass; browser syntax pass. Browser rerun на финальной сборке выполняет координатор перед новым independent review/push.
<!-- SECTION:NOTES:END -->
