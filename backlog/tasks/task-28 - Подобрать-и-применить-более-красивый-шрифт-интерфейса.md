---
id: TASK-28
title: Подобрать и применить более красивый шрифт интерфейса
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:21'
updated_date: '2026-10-18 21:35'
labels:
  - ui
  - typography
dependencies: []
references:
  - apps/web/src/lib/shared/presentation/styles/tokens.css
ordinal: 2026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец просит улучшить шрифт сайта. Сохранить компактность интерфейса и читаемость длинных подписей; перед выбором проверить текущую типографику и показать результат на реальных экранах.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Шрифт корректно отображает казахские Ә Ғ Қ Ң Ө Ұ Ү Һ І, русский, цифры и необходимые начертания без случайной подмены отдельных букв.
- [x] #2 Главная, каталог, меню, формы и читалка читаемы на 320/360/desktop в обеих темах; подписи и кнопки не обрезаются.
- [x] #3 Загрузка шрифта не блокирует чтение; есть подходящий запасной шрифт и учтена стоимость загрузки на слабой сети.
- [x] #4 Лицензия допускает использование на сайте; выбранный шрифт и источник записаны в карточке.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить Manrope из официального google/fonts: лицензию, реальные glyph cmap для kk/ru и вес файла. 2. Подключить один локальный variable WOFF2 (без внешних font-запросов и preload), сохранить системный fallback. 3. Браузерными проверками подтвердить загрузку собственного шрифта, fallback при ошибке, обе темы/локали и основные экраны. 4. Пересобрать Docker, прогнать регрессии и записать результаты в Review до remote CI/выпуска.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Manrope проверен по реальному cmap: отсутствуют ӘҒҚҢҰ и строчные, поэтому отклонён. Выбран Onest из официального https://github.com/google/fonts/tree/main/ofl/onest, variable wght 100–900. Полный файл конвертирован без удаления символов в WOFF2: 84056 байт. Все казахские буквы, русский алфавит и цифры подтверждены в cmap. SIL OFL 1.1 сохранена в static/fonts/onest-OFL.txt. Исходный TTF SHA256 966c5c29b4755da84b6854d5c21dd4eaa2420225d0e9874de602de176d4a9f31; WOFF2 SHA256 2d3db03b6507ed79fc0a51289f9032ec4685ea51bf53a61823ecbef886e9c943. Локальный font-face в CSS, hashed Vite asset, font-display swap, системный fallback, без preload и внешних сетевых запросов. Браузерная регрессия до подключения падала на отсутствии собственного шрифта. Финальная сборка и полный browser suite выполняются на обновляемом отдельном окружении.

Финально: Onest подключён; Docker build прошёл 90 frontend + 26 SSR тестов, app/preview 0 errors/warnings, architecture/contrast. Браузерная проверка подтвердила один same-origin запрос WOFF2 (84056 байт), загрузку regular/bold, fallback при отказе загрузки, главную/каталог/меню/регистрацию/реальный reader на 320/360/1280 в kk/ru × light/dark. Скриншоты осмотрены; переносов за пределы экрана в проверенных видах нет. Все 21 browser tests прошли дважды подряд на одной отдельной fixture. Логи tmp/font-build.log и tmp/integration/browser-repeat-{1,2}.log; скриншоты tmp/integration/final-screenshots/font-*.png. Общий Go/PostgreSQL/codec прогон: 73 test/subtest, 23 packages, без actual skips. Remote CI/выпуск и физические устройства не проверялись. Следующий шаг — наблюдаемый CI и приёмка тестового домена.

Код и тесты шрифта: 8722348. При проверке diff найдён пробел в исходном OFL.txt; в следующем служебном коммите убран только конечный пробел строки, текст лицензии сохранён.
<!-- SECTION:NOTES:END -->
