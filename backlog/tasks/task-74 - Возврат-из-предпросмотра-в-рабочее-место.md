---
id: TASK-74
title: Возврат из предпросмотра в рабочее место
status: Done
assignee:
  - '@studio-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 19:01'
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
- [x] #1 Возврат из предпросмотра восстанавливает проект, версию, рабочую секцию и позицию.
- [x] #2 Прямые ссылки, Back/reload и недоступная версия имеют честный fallback; приватность preview и account isolation сохраняются.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить контракт и добавить регрессии. 2. Реализовать согласованное изменение без новых правил и прав. 3. Выполнить проверки, записать SHA и ограничения приёмки.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
9c3bac5: preview URL несёт проект, ревизию, шаг и scroll; workspaceReturn принимает только точную локальную ссылку той же версии. Reader banner интегрируется владельцем Reader. Недоступная версия показывает явный fallback. Локально: svelte-check 0 ошибок/0 предупреждений; test:ui 46 pass; 9 domain/application тестов pass; frontend architecture pass; node --check studio-workflow.test.mjs pass. Browser studio-workflow.test.mjs/publishing-steps.test.mjs/project-cards.test.mjs и PostgreSQL integration на итоговом image ещё не выполнены; выпуск/независимая приёмка не подтверждены. Следующий шаг: общий rebuild и реальные browser/Go проверки координатором, затем Review; production SHA проверяет координатор.

Independent review P2 исправлен: 8414f46 сохраняет валидированный projectsReturn внутри workspace returnTo с запретом постороннего пути/схемы/account scope; df704ba синхронизирует выбранные revision/step через SvelteKit replaceState после завершения начального восстановления, не вызывая navigation/dirty confirm. Unit nested return RED→GREEN; browser full list→workspace→preview→return→all projects и выбор B/review→reload добавлены. Проверки после исправлений: svelte-check 0 errors/0 warnings; UI46 pass; 9 context/application тестов pass (54 assertions); architecture pass; browser syntax pass. Browser rerun на финальной сборке выполняет координатор перед новым independent review/push.

Локальная общая приёмка 20.10: полный Go/PostgreSQL -p1 ./... PASS (tmp/task72/go-final.log), 146 module +46UI PASS, Svelte0/0, architecture/contrast/ui-preview PASS. Browser studio-priority/studio-workflow/tag-search9/9 и publishing/cover/project-cards6/6 PASS; восстановление списка/скролла, preview→версия/шаг, reload, processing/grid/preflight подтверждены. Повторное независимое review3P2context исправлено, новых замечаний нет. Выпуск до push не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

CI6e2fa97 выявил test-only ожидание scroll до автоматической прокрутки ссылки Playwright. Исправлено: ссылка явно прокручивается в видимую область, фиксируется departure scroll; отдельно проверяются путь/версия/шаг/позиция/projectsReturn и строгая обратная ссылка. Studio-workflow4/4 local PASS; independent review без замечаний. SHA тестовых исправлений dd66757cb766a5f5809e018b85314174f0af2730. Требуется повторный Verify Qmanga; статус Review, выпуск не подтверждён.

Final remote verification 2026-10-20: SHA 4145910eb7bc649aaca2a2e890a44fbfa5f75575, Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39463685921 SUCCESS (both acceptance/browser shards, Go/PostgreSQL/codecs and frontend checks); Fast Release https://github.com/maqsatto/qmanga/actions/runs/39463685942 SUCCESS. Earlier CI follow-up failures are resolved. External production still reports 36b3a5f455eb37830408af44c2e2f00b48eabd51; deployment of this package is not confirmed, status remains Review. Physical Android TASK-97 remains not_run; new imports TASK-99–103 are queued only.

Приёмка и выпуск 21.10.2026: фактический production SHA e2c9c8910d6dedbe74eddbafa028b8bb67180340 повторно наблюдён во внешней HTTPS главной; /api/v1/health =200, postgres/s3 ok. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39474894775 SUCCESS на этом SHA: обе acceptance-половины,88 браузерных сценариев без failures, frontend/Go/PostgreSQL и ARM64 images. Актуальные CI-логи сверены с относящимися browser-регрессиями; независимый review и исправления описаны выше. Прежние записи «не запушено/не выпущено» исторические и заменены этим наблюдаемым выпуском. Критерии этой карточки выполнены; физический Android отдельно TASK-97 и здесь не заявляется.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Реализация и регрессии проверены, выпущены в production e2c9c89; фактический SHA/health и полный CI подтверждены. Физическая Android-приёмка отслеживается отдельно TASK-97.
<!-- SECTION:FINAL_SUMMARY:END -->
