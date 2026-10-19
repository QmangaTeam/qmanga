---
id: TASK-60
title: Пароль от 8 символов
status: Done
assignee:
  - '@codex'
created_date: '2026-10-19 01:42'
updated_date: '2026-10-19 02:00'
labels: []
dependencies: []
ordinal: 35026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец: «в пароле сделай 8 символов а не 10 а то бесит». Дополнительное описание: единый минимум8 в backend и формах регистрации/сброса; сохранить максимум128 и поддержку Unicode. Персональные данные и пароль в карточку не сохранять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Backend и frontend принимают8 и9 символов, отвергают7; подсказки обеих локалей согласованы
- [x] #2 Регрессии, review и выпуск проверены
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Backend identity+HTTP реальные PG регрессии успешны; frontend identity тесты успешны, включая границы7/8/9 и Unicode. Svelte check0errors/0warnings. Независимый final_owner_review существенных замечаний не нашёл. Выпуск ожидает CI/staging/prod общего SHA; персональные credentials не сохранялись в репозиторий.

Выпущен18eb4fdf02ee050b2cc2915195e03580a4a80594 на staging и production; внешний health/buildSHA проверен. CI39186717189 полныйsuccess; групповойbrowser50/50. Reviewбеззамечаний.
<!-- SECTION:NOTES:END -->
